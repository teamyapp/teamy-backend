package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Activator interface {
	FindActivatorByIDWithTx(ct context.Context, tx *transaction.Transaction, activatorID uint64) (entity.Activator, *errs.Error)
	FindActivatorByID(ct context.Context, activatorID uint64) (entity.Activator, *errs.Error)
	CreateActivator(ct context.Context, tx *transaction.Transaction, activator entity.Activator) *errs.Error
	UpdateActivator(ct context.Context, tx *transaction.Transaction, activator entity.Activator) *errs.Error
}
