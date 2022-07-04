package main

import (
	"database/sql"
	"log"
	"math/rand"
	"time"

	cloudConfig "github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/libs/runner"
	appsConfig "github.com/teamyapp/teamy-backend/apps/config"
	appsDep "github.com/teamyapp/teamy-backend/apps/dep"
	"github.com/teamyapp/teamy-backend/config"
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
	cloudAPIClientCfg, err := cloudConfig.CloudAPIClientFromEnv()
	if err != nil {
		log.Println(err)
		panic(err)
	}
	err = sqldb.Use(cfg.Config, func(sqlDB *sql.DB) error {
		err = sqldb.MigrateUp(sqlDB, "migrations")
		if err != nil {
			panic(err)
		}

		realTimeStateSyncer := dep.InitRealTimeStateSyncer(sqlDB)
		startServiceRunner(cloudAPIClientCfg, sqlDB, cfg.IdentityAPIEndpoint, realTimeStateSyncer)
		return nil
	})

	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func startServiceRunner(
	cloudAPIConfig cloudConfig.CloudAPIClient,
	sqlDB *sql.DB,
	identityAPIEndpoint string,
	realTimeStateSyncer *realtime.StateSyncer) {
	runnerConfig, err := runner.ServiceRunnerConfigFromEnv()
	if err != nil {
		panic(err)
	}

	githubCfg, err := appsConfig.GithubAppConfigFromEnv()
	if err != nil {
		panic(err)
	}

	githubApp, err := appsDep.InitGithubApp(cloudAPIConfig, githubCfg, sqlDB)
	if err != nil {
		panic(err)
	}

	teamyRealTimeStateSyncAPI := dep.InitRealTimeStateSyncAPI(identityAPIEndpoint, realTimeStateSyncer)
	teamyGraphQLAPI, err := dep.InitGraphQLAPI(identityAPIEndpoint, cloudAPIConfig, realTimeStateSyncer, sqlDB)
	if err != nil {
		panic(err)
	}
	rn := runner.NewServiceRunner(runnerConfig, []runner.Service{
		githubApp,
		teamyRealTimeStateSyncAPI,
		teamyGraphQLAPI,
	})
	rn.Start()
}
