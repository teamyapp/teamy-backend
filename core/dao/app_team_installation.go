package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppTeamInstallation interface {
	FindAppTeamInstallationsByAppID(ct context.Context, appID uint64) ([]entity.AppTeamInstallation, error)
	FindAppTeamInstallationsByTeamID(ct context.Context, teamID uint64) ([]entity.AppTeamInstallation, error)
	FindAppTeamInstallationByAppIDAndTeamID(ct context.Context, appID uint64, teamID uint64) (entity.AppTeamInstallation, error)
	CreateAppTeamInstallation(ct context.Context, appTeamInstallation entity.AppTeamInstallation) error
	UpdateAppTeamInstallation(ct context.Context, appTeamInstallation entity.AppTeamInstallation) error
	DeleteAppTeamInstallation(ct context.Context, appID uint64, teamID uint64) error
}
