package config

import (
	"time"

	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/app/dao/sqldb"
	"github.com/teamyapp/cloud/libs/errs"
)

type App struct {
	sqldb.Config
	config.Repo
	config.Service
	CloudWebAPIExternalBaseURL  string        `envconfig:"CLOUD_WEB_API_EXTERNAL_BASE_URL" default:"http://localhost:9011"`
	CloudWebAPIBaseURL          string        `envconfig:"CLOUD_WEB_API_BASE_URL" default:"http://localhost:9011"`
	CloudGRPCAPIHost            string        `envconfig:"CLOUD_GRPC_API_HOST" default:"localhost"`
	CloudGRPCAPIPort            int           `envconfig:"CLOUD_GRPC_API_PORT" default:"9011"`
	CloudGRPCAPIShouldEncrypt   bool          `envconfig:"CLOUD_GRPC_API_SHOULD_ENCRYPT" default:"false"`
	AppsServiceAccountAPIToken  string        `envconfig:"APPS_SERVICE_ACCOUNT_API_TOKEN" default:""`
	TeamyServiceAccountAPIToken string        `envconfig:"TEAMY_SERVICE_ACCOUNT_API_TOKEN" default:""`
	TeamyAPIHost                string        `envconfig:"TEAMY_API_HOST" default:"localhost"`
	TeamyAPIPort                int           `envconfig:"TEAMY_API_PORT" default:"9001"`
	TeamyAPIShouldEncrypt       bool          `envconfig:"TEAMY_API_SHOULD_ENCRYPT" default:"false"`
	RequestTimeout              time.Duration `envconfig:"REQUEST_TIMEOUT" default:"10s"`
	RequestRetryMaxCount        int           `envconfig:"REQUEST_RETRY_MAX_COUNT" default:"10"`
}

func AppFromEnv() (App, *errs.Error) {
	cfg := App{}
	err := config.FromEnv(&cfg)
	if err != nil {
		return App{}, err
	}

	return cfg, nil
}
