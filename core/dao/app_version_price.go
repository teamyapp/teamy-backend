package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppVersionPrice interface {
	FindAppVersionPricesByAppIDAndVersionNumberWithTx(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int) ([]entity.Money, *errs.Error)
	FindAppVersionPricesByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int) ([]entity.Money, *errs.Error)
	CreateAppVersionPrice(ct context.Context, tx *transaction.Transaction, appVersionPrice entity.AppVersionPrice) *errs.Error
	DeleteAppVersionPrice(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int) *errs.Error
}
