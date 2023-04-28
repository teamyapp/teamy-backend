package daov2

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppVersion interface {
	FindAppVersionsByAppID(ct context.Context, appID uint64) ([]entity.AppVersion, *errs.Error)
	FindAppVersionByAppIDAndVersionNumber(
	    ct context.Context, 
	    appID uint64,
		versionNumber int32,
	) (entity.AppVersion, *errs.Error)
	FindAppVersionByAppIDAndVersionNumberWithTx(
	    ct context.Context, 
	    tx *transaction.Transaction, 
	    appID uint64,
		versionNumber int32,
	) (entity.AppVersion, *errs.Error)
	FindAppVersionsByAppIDWithTx(
	    ct context.Context, 
	    tx *transaction.Transaction,
		appID uint64,
	) ([]entity.AppVersion, *errs.Error)
	FindMaxVersionNumberWithTx(ct context.Context, tx *transaction.Transaction, appID uint64) (int32, *errs.Error)
	CreateAppVersion(ct context.Context, tx *transaction.Transaction, appVersion entity.AppVersion) *errs.Error
	UpdateAppVersion(ct context.Context, tx *transaction.Transaction, appVersion entity.AppVersion) *errs.Error
	DeleteAppVersion(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int32) *errs.Error
}
