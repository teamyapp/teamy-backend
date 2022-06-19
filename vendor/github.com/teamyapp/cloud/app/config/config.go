package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"github.com/teamyapp/cloud/app/dao/sqldb"
)

type Repo struct {
	GitLongCommitHash string `envconfig:"GIT_LONG_COMMIT_HASH"`
	GitRepoOwner      string `envconfig:"GIT_REPO_OWNER"`
	GitRepoName       string `envconfig:"GIT_REPO_NAME"`
}

type App struct {
	Repo
	sqldb.Config
	AccessTokenTTL     time.Duration `envconfig:"ACCESS_TOKEN_TTL" default:""`
	GoogleClientID     string        `envconfig:"GOOGLE_CLIENT_ID" default:""`
	GoogleClientSecret string        `envconfig:"GOOGLE_CLIENT_SECRET" default:""`
	GitHubClientID     string        `envconfig:"GITHUB_CLIENT_ID" default:""`
	GitHubClientSecret string        `envconfig:"GITHUB_CLIENT_SECRET" default:""`
	JWTSigningKey      string        `envconfig:"JWT_SIGNING_KEY" default:""`
	GenRangeSize       int           `envconfig:"GEN_RANGE_SIZE" default:"100"`
	WebAPIBaseURL      string        `envconfig:"WEB_API_BASE_URL" default:""`
}

type CloudAPIClient struct {
	Host          string `envconfig:"CLOUD_API_HOST" default:"localhost"`
	Port          int    `envconfig:"CLOUD_API_PORT" default:"9501"`
	ShouldEncrypt bool   `envconfig:"CLOUD_API_SHOULD_ENCRYPT" default:"false"`
}

func AppFromEnv() (App, error) {
	cfg := App{}
	err := FromEnv(&cfg)
	if err != nil {
		log.Println(err)
		return App{}, err
	}
	return cfg, nil
}

func CloudAPIClientFromEnv() (CloudAPIClient, error) {
	cfg := CloudAPIClient{}
	err := FromEnv(&cfg)
	if err != nil {
		log.Println(err)
		return CloudAPIClient{}, err
	}
	return cfg, nil
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
