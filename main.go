package main

import (
	"database/sql"
	"log"

	"github.com/teamyapp/one/db"
	"github.com/teamyapp/teamy-backend/app/api/gql"
	"github.com/teamyapp/teamy-backend/app/config"
	"github.com/teamyapp/teamy-backend/app/dep"
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

	panic(db.With(cfg.OneConfig, func(sqlDB *sql.DB) error {
		gqlResolver := dep.InitGraphQLResolver(sqlDB)
		server, err := gql.NewServer(cfg.IdentityAPIEndpoint, gqlResolver, cfg.GraphQLAPIPort)
		if err != nil {
			panic(err)
		}

		log.Printf("GraphQL server started at %d\n", cfg.GraphQLAPIPort)
		panic(server.ListenAndServe())
		return nil
	}))
}
