package config

import (
	"log"

	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dao/sqldb"
)

type Config struct {
	sqldb.Config
	config.RepoConfig
	IdentityAPIEndpoint string `envconfig:"IDENTITY_API_ENDPOINT" default:"http://localhost:9500/identity"`
	GraphQLAPIPort      int    `envconfig:"GRAPH_QL_API_PORT" default:"9000"`
	GraphQLAPIV2Port    int    `envconfig:"GRAPH_QL_API_V2_PORT" default:"9001"`
}

func FromEnv() (Config, error) {
	cfg := Config{}
	err := config.FromEnv(&cfg)
	if err != nil {
		log.Println(err)
		return Config{}, err
	}
	return cfg, nil
}
