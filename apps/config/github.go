package config

import (
	"log"
	"time"

	"github.com/teamyapp/cloud/app/config"
)

type GithubAppConfig struct {
	AppName                   string        `envconfig:"APPS_GITHUB_APP_NAME"`
	InstallationValidDuration time.Duration `envconfig:"APPS_GITHUB_INSTALLATION_VALID_DURATION" default:"5m"`
	WebhookSecret             string        `envconfig:"APPS_GITHUB_WEBHOOK_SECRET"`
}

func GithubAppConfigFromEnv() (GithubAppConfig, error) {
	cfg := GithubAppConfig{}
	err := config.FromEnv(&cfg)
	if err != nil {
		log.Println(err)
		return GithubAppConfig{}, err
	}

	return cfg, nil
}
