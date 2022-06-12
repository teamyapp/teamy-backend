package dao

import (
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubAppInstallation interface {
	CreateGithubAppInstallation(installation entity.GithubAppInstallation) error
}
