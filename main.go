package main

import (
	"database/sql"
	"log"
	"math/rand"
	"sync"
	"time"

	cloudConfig "github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/teamy-backend/apps"
	appsConfig "github.com/teamyapp/teamy-backend/apps/config"
	appsDep "github.com/teamyapp/teamy-backend/apps/dep"
	"github.com/teamyapp/teamy-backend/config"
	"github.com/teamyapp/teamy-backend/core/api/gql"
	"github.com/teamyapp/teamy-backend/core/dep"
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

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			startGraphQLServer(cfg, cloudAPIClientCfg, sqlDB)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			startAppRunner(cloudAPIClientCfg, sqlDB)
		}()
		wg.Wait()
		return nil
	})

	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func startGraphQLServer(cfg config.Config, cloudAPIConfig cloudConfig.CloudAPIClient, sqlDB *sql.DB) {
	gqlResolver, err := dep.InitGraphQLResolver(sqlDB, cloudAPIConfig)
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

func startAppRunner(cloudAPIConfig cloudConfig.CloudAPIClient, sqlDB *sql.DB) {
	runnerConfig, err := appsConfig.AppRunnerConfigFromEnv()
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

	runner := apps.NewAppRunner(runnerConfig, []apps.App{
		githubApp,
	})
	err = runner.Start()
	if err != nil {
		panic(err)
	}
}
