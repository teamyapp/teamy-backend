package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao/entity"
)

type MaxViewersActivator interface {
	FindMaxViewersActivatorByIDWithTx(ct context.Context, tx *transaction.Transaction, activatorID uint64) (entity.PartialMaxViewersActivator, *errs.Error)
	FindMaxViewersActivatorByID(ct context.Context, activatorID uint64) (entity.PartialMaxViewersActivator, *errs.Error)
	CreateMaxViewersActivator(ct context.Context, tx *transaction.Transaction, activatorID uint64, activator entity.PartialMaxViewersActivator) *errs.Error
	UpdateMaxViewersActivator(ct context.Context, tx *transaction.Transaction, activatorID uint64, activator entity.PartialMaxViewersActivator) *errs.Error
	DeleteMaxViewersActivator(ct context.Context, tx *transaction.Transaction, activatorID uint64) *errs.Error
}
