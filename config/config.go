package config

import (
	"log"

	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dao/sqldb"
)

type Config struct {
	sqldb.Config
	config.Repo
	CloudWebAPIBaseURL         string `envconfig:"CLOUD_WEB_API_BASE_URL" default:"http://localhost:9011"`
	CloudGRPCAPIHost           string `envconfig:"CLOUD_GRPC_API_HOST" default:"localhost"`
	CloudGRPCAPIPort           int    `envconfig:"CLOUD_GRPC_API_PORT" default:"9011"`
	CloudGRPCAPIShouldEncrypt  bool   `envconfig:"CLOUD_GRPC_API_SHOULD_ENCRYPT" default:"false"`
	AppsServiceAccountAPIToken string `envconfig:"APPS_SERVICE_ACCOUNT_API_TOKEN" default:""`
	TeamyAPIHost               string `envconfig:"TEAMY_API_HOST" default:"localhost"`
	TeamyAPIPort               int    `envconfig:"TEAMY_API_PORT" default:"9001"`
	TeamyAPIShouldEncrypt      bool   `envconfig:"TEAMY_API_SHOULD_ENCRYPT" default:"false"`
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
