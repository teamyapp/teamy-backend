package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppVersionChange interface {
	FindAppVersionChangesByAppIDAndVersionNumberWithTx(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int) ([]string, *errs.Error)
	FindAppVersionChangesByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int) ([]string, *errs.Error)
	CreateAppVersionChange(ct context.Context, tx *transaction.Transaction, appVersionChange entity.AppVersionChange) *errs.Error
}
