package github

import (
	"time"

	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/libs/errs"
)

type AppConfig struct {
	AppName                   string        `envconfig:"APPS_GITHUB_APP_NAME"`
	InstallationValidDuration time.Duration `envconfig:"APPS_GITHUB_INSTALLATION_VALID_DURATION" default:"5m"`
	WebhookSecret             string        `envconfig:"APPS_GITHUB_WEBHOOK_SECRET"`
}

func AppConfigFromEnv() (AppConfig, *errs.Error) {
	cfg := AppConfig{}
	err := config.FromEnv(&cfg)
	if err != nil {
		return AppConfig{}, &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
	}

	return cfg, nil
}
