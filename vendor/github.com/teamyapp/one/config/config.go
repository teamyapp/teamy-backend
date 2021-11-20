package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	DBHost     string `envconfig:"DB_HOST" default:"localhost"`
	DBPort     int    `envconfig:"DB_PORT" default:"5432"`
	DBUser     string `envconfig:"DB_USER"`
	DBName     string `envconfig:"DB_NAME" default:"teamy"`
	DBPassword string `envconfig:"DB_PASSWORD"`
	DBSSLMode  string `envconfig:"DB_SSL_MODE" default:"require"`
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
	err := autoLoadEnv()
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

func autoLoadEnv() error {
	_, err := os.Stat(".env")
	if os.IsNotExist(err) {
		return nil
	}

	return godotenv.Load()
}
