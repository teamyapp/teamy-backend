package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppVersion interface {
	FindAppVersionByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int32) (entity.AppVersion, *errs.Error)
	FindAppVersionsByAppID(ct context.Context, appID uint64) ([]entity.AppVersion, *errs.Error)
	CreateAppVersion(ct context.Context, appVersion entity.AppVersion) (int32, *errs.Error)
	UpdateAppVersion(ct context.Context, appVersion entity.AppVersion) *errs.Error
	DeleteAppVersion(ct context.Context, appID uint64, versionNumber int32) *errs.Error
}
