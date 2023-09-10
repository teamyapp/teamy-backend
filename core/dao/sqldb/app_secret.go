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

func (*AppSecret) FindAppSecretByIDWithTx(ct context.Context, tx *transaction.Transaction, appSecretID uint64) (entity.AppSecret, *errs.Error) {
	panic("unimplemented")
}

func (*AppSecret) FindSecretsByAppID(ct context.Context, appID uint64) ([]entity.AppSecret, *errs.Error) {
	panic("unimplemented")
}

func (*AppSecret) CreateAppSecret(ct context.Context, appSecret entity.AppSecret) (entity.AppSecret, *errs.Error) {
	panic("unimplemented")
}

func (*AppSecret) UpdateAppSecretWithTx(ct context.Context, tx *transaction.Transaction, appSecretID uint64, appSecret entity.AppSecret) *errs.Error {
	panic("unimplemented")
}

func (*AppSecret) DeleteAppSecretWithTx(ct context.Context, tx *transaction.Transaction, appSecretID uint64) *errs.Error {
	panic("unimplemented")
}

var _ dao.AppSecret = (*AppSecret)(nil)

func NewAppSecret() *AppSecret {
	return &AppSecret{}
}
