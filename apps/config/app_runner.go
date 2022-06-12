package config

import (
	"log"

	"github.com/teamyapp/cloud/app/config"
)

type AppRunnerConfig struct {
	WebAPIPort int `envconfig:"APP_RUNNER_WEB_API_PORT" default:"9001"`
}

func AppRunnerConfigFromEnv() (AppRunnerConfig, error) {
	cfg := AppRunnerConfig{}
	err := config.FromEnv(&cfg)
	if err != nil {
		log.Println(err)
		return AppRunnerConfig{}, err
	}
	return cfg, nil
}
