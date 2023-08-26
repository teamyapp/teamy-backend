package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppVersion interface {
	FindAppVersionsByAppIDWithTx(ct context.Context, tx *transaction.Transaction, appID uint64) ([]entity.AppVersion, *errs.Error)
	FindAppVersionsByAppID(ct context.Context, appID uint64) ([]entity.AppVersion, *errs.Error)
	FindMaxVersionNumberWithTx(ct context.Context, tx *transaction.Transaction, appID uint64) (int, *errs.Error)
	CreateAppVersion(ct context.Context, tx *transaction.Transaction, appVersion entity.AppVersion) *errs.Error
	DeleteAppVersion(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int) *errs.Error
	FindAppVersionByAppIDAndVersionNumberWithTx(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int) (entity.AppVersion, *errs.Error)
}
