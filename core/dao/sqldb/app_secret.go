package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppSecret struct {
}

// CreateAppSecret implements dao.AppSecret.
func (*AppSecret) CreateAppSecret(ct context.Context, appSecret entity.AppSecret) (entity.AppSecret, *errs.Error) {
	panic("unimplemented")
}

// DeleteAppSecretWithTx implements dao.AppSecret.
func (*AppSecret) DeleteAppSecretWithTx(ct context.Context, tx *transaction.Transaction, appSecretID uint64) *errs.Error {
	panic("unimplemented")
}

// FindAppSecretByIDWithTx implements dao.AppSecret.
func (*AppSecret) FindAppSecretByIDWithTx(ct context.Context, tx *transaction.Transaction, appSecretID uint64) (entity.AppSecret, *errs.Error) {
	panic("unimplemented")
}

// FindSecretsByAppID implements dao.AppSecret.
func (*AppSecret) FindSecretsByAppID(ct context.Context, appID uint64) ([]entity.AppSecret, *errs.Error) {
	panic("unimplemented")
}

// UpdateAppSecretWithTx implements dao.AppSecret.
func (*AppSecret) UpdateAppSecretWithTx(ct context.Context, tx *transaction.Transaction, appSecretID uint64, appSecret entity.AppSecret) *errs.Error {
	panic("unimplemented")
}

var _ dao.AppSecret = (*AppSecret)(nil)

func NewAppSecret() *AppSecret {
	return &AppSecret{}
}
