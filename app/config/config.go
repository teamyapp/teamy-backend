package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DbHost         string `envconfig:"DB_HOST" default:"localhost"`
	DbPort         int    `envconfig:"DB_PORT" default:"5432"`
	DbUser         string `envconfig:"DB_USER"`
	DBName         string `envconfig:"DB_NAME" default:"teamy"`
	DbPassword     string `envconfig:"DB_PASSWORD"`
	GraphQLAPIPort int    `envconfig:"GRAPH_QL_API_PORT" default:"9000"`
}

func FromEnv() (Config, error) {
	err := autoLoadEnv()
	if err != nil {
		log.Println(err)
		return Config{}, err
	}

	config := Config{}
	err = envconfig.Process("", &config)
	if err != nil {
		log.Println(err)
		return Config{}, err
	}
	return config, nil
}

func autoLoadEnv() error {
	_, err := os.Stat(".env")
	if os.IsNotExist(err) {
		return nil
	}

	return godotenv.Load()
}
