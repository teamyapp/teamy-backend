package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppSecret interface {
	FindAppSecretByIDWithTx(ct context.Context, tx *transaction.Transaction, appSecretID uint64) (entity.AppSecret, *errs.Error)
	FindAppSecretsByAppIDWithTx(ct context.Context, tx *transaction.Transaction, appID uint64) ([]entity.AppSecret, *errs.Error)
	FindSecretsByAppID(ct context.Context, appID uint64) ([]entity.AppSecret, *errs.Error)
	CreateAppSecret(ct context.Context, tx *transaction.Transaction, appSecret entity.AppSecret) *errs.Error
	UpdateAppSecret(ct context.Context, tx *transaction.Transaction, appSecretID uint64, appSecret entity.AppSecret) *errs.Error
	DeleteAppSecret(ct context.Context, tx *transaction.Transaction, appSecretID uint64) *errs.Error
}
