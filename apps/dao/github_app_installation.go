package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubAppInstallation interface {
	FindInstallationByID(ct context.Context, installationID uint64) (entity.GithubAppInstallation, error)
	CreateGithubAppInstallation(ct context.Context, installation entity.GithubAppInstallation) error
}
