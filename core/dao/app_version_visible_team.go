package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppVersionVisibleTeam interface {
	FindAppVersionVisibleTeam(ct context.Context, appID uint64, versionNumber int32, teamID uint64) (entity.AppVersionVisibleTeam, *errs.Error)
	FindAppVersionVisibleTeamsByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int32) ([]entity.AppVersionVisibleTeam, *errs.Error)
	FindAppVersionVisibleTeamsByTeamID(ct context.Context, teamID uint64) ([]entity.AppVersionVisibleTeam, *errs.Error)
	CreateAppVersionVisibleTeam(ct context.Context, appVersionVisibleTeam entity.AppVersionVisibleTeam) *errs.Error
	DeleteAppVersionVisibleTeam(ct context.Context, appID uint64, versionNumber int32, teamID uint64) *errs.Error
	DeleteAppVersionVisibleTeamByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int32) *errs.Error
}
