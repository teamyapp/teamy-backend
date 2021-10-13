package main

import (
	"log"

	"github.com/teamyapp/teamy-backend/app/api/gql"
	"github.com/teamyapp/teamy-backend/app/config"
	"github.com/teamyapp/teamy-backend/app/dep"
	"github.com/teamyapp/teamy-backend/app/repo/db"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
}

func main() {
	cfg, err := config.FromEnv()
	if err != nil {
		log.Println(err)
		panic(err)
	}

	sqlDB, err := db.Connect(cfg)
	if err != nil {
		panic(err)
	}
	defer sqlDB.Close()

	err = sqlDB.Ping()
	if err != nil {
		panic(err)
	}

	executionService := dep.InitExecutionService(sqlDB)
	server, err := gql.NewServer(executionService, cfg.GraphQLAPIPort)
	if err != nil {
		panic(err)
	}

	log.Printf("GraphQL server started at %d\n", cfg.GraphQLAPIPort)
	panic(server.ListenAndServe())
}
