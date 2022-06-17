package main

import (
	"database/sql"
	"log"
	"math/rand"
	"sync"
	"time"

	cloudConfig "github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	appsConfig "github.com/teamyapp/teamy-backend/apps/config"
	appsDep "github.com/teamyapp/teamy-backend/apps/dep"
	"github.com/teamyapp/teamy-backend/config"
	"github.com/teamyapp/teamy-backend/core/api/gql"
	"github.com/teamyapp/teamy-backend/core/dep"
	"github.com/teamyapp/teamy-backend/infras/runner"
	"github.com/teamyapp/teamy-backend/infras/service"
	"github.com/teamyapp/teamy-backend/infras/storage"
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

		realTimeCollections := service.NewRealTimeCollections()
		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			startGraphQLServer(
				cfg,
				cloudAPIClientCfg,
				sqlDB,
				realTimeCollections.InMemoryRealTimeCollections())
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			startServiceRunner(cloudAPIClientCfg, sqlDB, realTimeCollections)
		}()
		wg.Wait()
		return nil
	})

	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func startGraphQLServer(
	cfg config.Config,
	cloudAPIConfig cloudConfig.CloudAPIClient,
	sqlDB *sql.DB,
	rtCollections *storage.RealTimeCollections) {
	gqlResolver, err := dep.InitGraphQLResolver(sqlDB, cloudAPIConfig, rtCollections)
	if err != nil {
		panic(err)
	}

	server, err := gql.NewServer(cfg.IdentityAPIEndpoint, gqlResolver, cfg.CoreGraphQLAPIPort)
	if err != nil {
		panic(err)
	}

	log.Printf("GraphQL server started at %d\n", cfg.CoreGraphQLAPIPort)
	panic(server.ListenAndServe())
}

func startServiceRunner(
	cloudAPIConfig cloudConfig.CloudAPIClient,
	sqlDB *sql.DB,
	rtCollections service.RealTimeCollections) {
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

	rn := runner.NewServiceRunner(runnerConfig, []runner.Service{
		githubApp,
		rtCollections,
	})
	err = rn.Start()
	if err != nil {
		panic(err)
	}
}
