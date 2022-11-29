package main

import (
	"database/sql"
	"fmt"
	"math/rand"
	"os"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/cloud/libs/retry"
	"github.com/teamyapp/cloud/libs/retry/backoff"
	"github.com/teamyapp/cloud/libs/rpc"
	"github.com/teamyapp/cloud/libs/runner"
	"github.com/teamyapp/cloud/libs/runtime"
	appsDep "github.com/teamyapp/teamy-backend/apps/dep"
	"github.com/teamyapp/teamy-backend/apps/github"
	appsDi "github.com/teamyapp/teamy-backend/apps/inject"
	"github.com/teamyapp/teamy-backend/config"
	"github.com/teamyapp/teamy-backend/core/api"
	"github.com/teamyapp/teamy-backend/core/dep"
	"github.com/teamyapp/teamy-backend/core/inject"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func main() {
	visibleLevel := obs.LogLevel(getEnv("LOG_VISIBLE_SEVERITY", "Info"))
	dataCollector := dep.InitDataCollector("teamy/backend", visibleLevel)
	inject.Injector.BindType(new(obs.DataCollector), func() interface{} {
		return dataCollector
	})
	appsDi.Injector.BindType(new(obs.DataCollector), func() interface{} {
		return dataCollector
	})

	cfg, err := config.FromEnv(dataCollector)
	if err != nil {
		dataCollector.Logger.Log(obs.Fatal, obs.Props{obs.CauseProp: err})
		panic(err)
	}

	gitCommitLink := fmt.Sprintf("https://github.com/%s/%s/commit/%s",
		cfg.GitRepoOwner,
		cfg.GitRepoName,
		cfg.GitLongCommitHash)
	dataCollector.Logger.Log(obs.Info, obs.Props{
		obs.MessageProp: map[string]interface{}{
			"gitCommitLink": gitCommitLink,
		},
	})
	err = sqldb.Use(dataCollector, cfg.Config, func(sqlDB *sql.DB) error {
		err = sqldb.MigrateUp(dataCollector, sqlDB, "migrations", 0)
		if err != nil {
			dataCollector.Logger.Log(obs.Fatal, obs.Props{obs.CauseProp: err})
			panic(err)
		}

		realTimeStateSyncer := dep.InitRealTimeStateSyncer(dataCollector, sqlDB)
		return startServiceRunner(dataCollector, cfg, sqlDB, realTimeStateSyncer)
	})

	if err != nil {
		dataCollector.Logger.Log(obs.Fatal, obs.Props{obs.CauseProp: err})
		panic(err)
	}
}

func startServiceRunner(
	dataCollector obs.DataCollector,
	cfg config.Config,
	sqlDB *sql.DB,
	realTimeStateSyncer *realtime.StateSyncer,
) error {
	runnerConfig, err := runner.ServiceRunnerConfigFromEnv(dataCollector)
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	githubCfg, err := github.AppConfigFromEnv(dataCollector)
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	exponentialBackOff := backoff.NewExponentialBuilder().Build()
	maxCountRetry := retry.NewMaxCount(runtime.NewBuiltInRuntime(), exponentialBackOff, cfg.RequestRetryMaxCount)
	cloudClientRegistry, err := cloudAPI.NewClientRegistry(
		dataCollector,
		rpc.ConnectionConfig{
			Host:          cfg.CloudGRPCAPIHost,
			Port:          cfg.CloudGRPCAPIPort,
			ShouldEncrypt: cfg.CloudGRPCAPIShouldEncrypt,
			GetAccessToken: func() string {
				return cfg.TeamyServiceAccountAPIToken
			},
			RequestTimeout: cfg.RequestTimeout,
		}, maxCountRetry)
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	teamyClientRegistry, err := api.NewClientRegistry(
		dataCollector,
		rpc.ConnectionConfig{
			Host:          cfg.TeamyAPIHost,
			Port:          cfg.TeamyAPIPort,
			ShouldEncrypt: cfg.TeamyAPIShouldEncrypt,
			GetAccessToken: func() string {
				return cfg.AppsServiceAccountAPIToken
			},
			RequestTimeout: cfg.RequestTimeout,
		}, maxCountRetry)
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	githubApp, err := appsDep.InitGithubApp(
		dataCollector,
		cloudClientRegistry,
		teamyClientRegistry,
		githubCfg,
		sqlDB)
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	realTimeStateSyncAPI := dep.InitRealTimeStateSyncAPI(
		dataCollector,
		realTimeStateSyncer)
	graphQLAPI, err := dep.InitGraphQLAPI(
		dataCollector,
		dep.CloudWebAPIExternalBaseURL(cfg.CloudWebAPIExternalBaseURL),
		cloudClientRegistry,
		realTimeStateSyncer,
		sqlDB)
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	taskRPCAPI := dep.InitTaskRPCAPI(
		dataCollector,
		cloudClientRegistry,
		realTimeStateSyncer,
		sqlDB)
	sprintRPCAPI := dep.InitSprintRPCAPI(
		dataCollector,
		cloudClientRegistry,
		realTimeStateSyncer,
		sqlDB)
	rn := runner.NewServiceRunner(
		dataCollector,
		runnerConfig, []runner.Service{
			githubApp,
			graphQLAPI,
			realTimeStateSyncAPI,
			taskRPCAPI,
			sprintRPCAPI,
		})
	rn.Start()
	return nil
}

func getEnv(name string, defaultVal string) string {
	value := os.Getenv(name)
	if len(value) > 0 {
		return value
	}

	return defaultVal
}
