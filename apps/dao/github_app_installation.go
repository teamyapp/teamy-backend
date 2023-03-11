package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubAppInstallation interface {
	FindInstallationIDByTeamID(ct context.Context, teamID uint64) (int, *errs.Error)
	FindInstallationByID(ct context.Context, installationID uint64) (entity.GithubAppInstallation, *errs.Error)
	CreateGithubAppInstallation(ct context.Context, installation entity.GithubAppInstallation) *errs.Error
}
