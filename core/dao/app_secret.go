package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppSecret interface {
	FindAppSecretByIDWithTx(ct context.Context, tx *transaction.Transaction, appSecretID uint64) (entity.AppSecret, *errs.Error)
	FindSecretsByAppID(ct context.Context, appID uint64) ([]entity.AppSecret, *errs.Error)
	CreateAppSecret(ct context.Context, appSecret entity.AppSecret) (entity.AppSecret, *errs.Error)
	UpdateAppSecretWithTx(ct context.Context, tx *transaction.Transaction, appSecretID uint64, appSecret entity.AppSecret) *errs.Error
	DeleteAppSecretWithTx(ct context.Context, tx *transaction.Transaction, appSecretID uint64) *errs.Error
}
