package main

import (
	"database/sql"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/teamy-backend/app/api/gql"
	"github.com/teamyapp/teamy-backend/app/api/gqlv2"
	"github.com/teamyapp/teamy-backend/app/config"
	"github.com/teamyapp/teamy-backend/app/dep"
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
		err = sqldb.MigrateUp(sqlDB, sqldb.DefaultMigrationRoot)
		if err != nil {
			log.Println(err)
			return err
		}

		wg := sync.WaitGroup{}
		wg.Add(1)
		go func() {
			defer wg.Done()
			StartGraphQLServer(cfg, sqlDB)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			StartGraphQLV2Server(cfg, sqlDB)
		}()
		wg.Wait()
		return nil
	})

	if err != nil {
		log.Println(err)
		panic(err)
	}
}

func StartGraphQLServer(cfg config.Config, sqlDB *sql.DB) {
	gqlResolver := dep.InitGraphQLResolver(sqlDB)
	server, err := gql.NewServer(cfg.IdentityAPIEndpoint, gqlResolver, cfg.GraphQLAPIPort)
	if err != nil {
		panic(err)
	}

	log.Printf("GraphQL server started at %d\n", cfg.GraphQLAPIPort)
	panic(server.ListenAndServe())
}

func StartGraphQLV2Server(cfg config.Config, sqlDB *sql.DB) {
	gqlV2Resolver := dep.InitGraphQLV2Resolver(sqlDB)
	server, err := gqlv2.NewServer(cfg.IdentityAPIEndpoint, gqlV2Resolver, cfg.GraphQLAPIV2Port)
	if err != nil {
		panic(err)
	}

	log.Printf("GraphQL V2 server started at %d\n", cfg.GraphQLAPIV2Port)
	panic(server.ListenAndServe())
}
