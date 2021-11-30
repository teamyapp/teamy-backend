package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DBHost            string `envconfig:"DB_HOST" default:"localhost"`
	DBPort            int    `envconfig:"DB_PORT" default:"5432"`
	DBUser            string `envconfig:"DB_USER"`
	DBName            string `envconfig:"DB_NAME" default:"teamy"`
	DBPassword        string `envconfig:"DB_PASSWORD"`
	DBSSLMode         string `envconfig:"DB_SSL_MODE" default:"require"`
	GitLongCommitHash string `envconfig:"GIT_LONG_COMMIT_HASH"`
	RepoOwner         string `envconfig:"REPO_OWNER"`
	RepoName          string `envconfig:"REPO_NAME"`
}

func OneConfigFromEnv() (Config, error) {
	config := Config{}
	err := FromEnv(&config)
	if err != nil {
		log.Println(err)
		return Config{}, err
	}
	return config, nil
}

func FromEnv(config interface{}) error {
	err := autoLoadEnv(".env")
	if err != nil {
		log.Println(err)
		return err
	}

	err = autoLoadEnv(".repo.env")
	if err != nil {
		log.Println(err)
		return err
	}

	err = envconfig.Process("", config)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func autoLoadEnv(fileName string) error {
	_, err := os.Stat(fileName)
	if err == nil {
		return godotenv.Load(fileName)
	} else if os.IsNotExist(err) {
		return nil
	} else {
		return err
	}
}
