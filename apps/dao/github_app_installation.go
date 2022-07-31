package dao

import (
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubAppInstallation interface {
	FindInstallationByID(installationID uint64) (entity.GithubAppInstallation, error)
	CreateGithubAppInstallation(installation entity.GithubAppInstallation) error
}
