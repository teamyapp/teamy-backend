package github

import (
	"time"

	"github.com/teamyapp/cloud/app/config"
	"github.com/teamyapp/cloud/libs/obs"
)

type AppConfig struct {
	AppName                   string        `envconfig:"APPS_GITHUB_APP_NAME"`
	InstallationValidDuration time.Duration `envconfig:"APPS_GITHUB_INSTALLATION_VALID_DURATION" default:"5m"`
	WebhookSecret             string        `envconfig:"APPS_GITHUB_WEBHOOK_SECRET"`
}

func AppConfigFromEnv(dataCollector obs.DataCollector) (AppConfig, error) {
	cfg := AppConfig{}
	err := config.FromEnv(dataCollector, &cfg)
	if err != nil {
		dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return AppConfig{}, err
	}

	return cfg, nil
}
