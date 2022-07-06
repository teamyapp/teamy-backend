package github

import (
	"log"
	"time"

	"github.com/teamyapp/cloud/app/config"
)

type AppConfig struct {
	AppName                   string        `envconfig:"APPS_GITHUB_APP_NAME"`
	InstallationValidDuration time.Duration `envconfig:"APPS_GITHUB_INSTALLATION_VALID_DURATION" default:"5m"`
	WebhookSecret             string        `envconfig:"APPS_GITHUB_WEBHOOK_SECRET"`
}

func AppConfigFromEnv() (AppConfig, error) {
	cfg := AppConfig{}
	err := config.FromEnv(&cfg)
	if err != nil {
		log.Println(err)
		return AppConfig{}, err
	}

	return cfg, nil
}
