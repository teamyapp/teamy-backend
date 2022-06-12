package config

import (
	"log"

	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dao/sqldb"
)

type CloudAPIConfig struct {
	Host          string
	Port          int
	ShouldEncrypt bool
}

type Config struct {
	sqldb.Config
	config.RepoConfig
	IdentityAPIEndpoint   string `envconfig:"IDENTITY_API_ENDPOINT" default:"http://localhost:9500/identity"`
	CoreGraphQLAPIPort    int    `envconfig:"CORE_GRAPH_QL_API_PORT" default:"9000"`
	CoreWebAPIPort        int    `envconfig:"CORE_WEB_API_PORT" default:"9001"`
	CloudAPIHost          string `envconfig:"CLOUD_API_HOST" default:"localhost"`
	CloudAPIPort          int    `envconfig:"CLOUD_API_PORT" default:"9501"`
	CloudAPIShouldEncrypt bool   `envconfig:"CLOUD_API_SHOULD_ENCRYPT" default:"false"`
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
