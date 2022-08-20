package main

import (
	"database/sql"
	"log"
	"math/rand"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/libs/rpc"
	"github.com/teamyapp/cloud/libs/runner"
	appsDep "github.com/teamyapp/teamy-backend/apps/dep"
	"github.com/teamyapp/teamy-backend/apps/github"
	"github.com/teamyapp/teamy-backend/config"
	"github.com/teamyapp/teamy-backend/core/api"
	"github.com/teamyapp/teamy-backend/core/dep"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	rand.Seed(time.Now().UnixNano())
}

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Println(err)
		panic(err)
	}

	log.Printf(
		"Git Commit: https://github.com/%s/%s/commit/%s\n",
		cfg.GitRepoOwner,
		cfg.GitRepoName,
		cfg.GitLongCommitHash)
	err = sqldb.Use(cfg.Config, func(sqlDB *sql.DB) error {
		err = sqldb.MigrateUp(sqlDB, "migrations", 0)
		if err != nil {
			panic(err)
		}

		realTimeStateSyncer := dep.InitRealTimeStateSyncer(sqlDB)
		startServiceRunner(cfg, sqlDB, realTimeStateSyncer)
		return nil
	})

	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func startServiceRunner(
	cfg config.Config,
	sqlDB *sql.DB,
	realTimeStateSyncer *realtime.StateSyncer) {
	runnerConfig, err := runner.ServiceRunnerConfigFromEnv()
	if err != nil {
		panic(err)
	}

	githubCfg, err := github.AppConfigFromEnv()
	if err != nil {
		panic(err)
	}
	cloudClientRegistry, err := cloudAPI.NewClientRegistry(rpc.ConnectionConfig{
		Host:          cfg.CloudGRPCAPIHost,
		Port:          cfg.CloudGRPCAPIPort,
		ShouldEncrypt: cfg.CloudGRPCAPIShouldEncrypt,
	})
	if err != nil {
		panic(err)
	}
	teamyClientRegistry, err := api.NewClientRegistry(rpc.ConnectionConfig{
		Host:          cfg.TeamyAPIHost,
		Port:          cfg.TeamyAPIPort,
		ShouldEncrypt: cfg.TeamyAPIShouldEncrypt,
		GetAccessToken: func() string {
			return cfg.AppsServiceAccountAPIToken
		},
	})
	if err != nil {
		panic(err)
	}

	githubApp, err := appsDep.InitGithubApp(cloudClientRegistry, teamyClientRegistry, githubCfg, sqlDB)
	if err != nil {
		panic(err)
	}

	realTimeStateSyncAPI := dep.InitRealTimeStateSyncAPI(realTimeStateSyncer)
	graphQLAPI, err := dep.InitGraphQLAPI(
		dep.CloudWebAPIExternalBaseURL(cfg.CloudWebAPIExternalBaseURL),
		cloudClientRegistry,
		realTimeStateSyncer,
		sqlDB)
	if err != nil {
		panic(err)
	}

	taskRPCAPI := dep.InitTaskRPCAPI(cloudClientRegistry, realTimeStateSyncer, sqlDB)
	sprintRPCAPI := dep.InitSprintRPCAPI(cloudClientRegistry, realTimeStateSyncer, sqlDB)
	rn := runner.NewServiceRunner(runnerConfig, []runner.Service{
		githubApp,
		graphQLAPI,
		realTimeStateSyncAPI,
		taskRPCAPI,
		sprintRPCAPI,
	})
	rn.Start()
}
