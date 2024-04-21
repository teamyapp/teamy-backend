package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao/entity"
)

type PercentageActivator interface {
	FindPercentageActivatorByIDWithTx(ct context.Context, tx *transaction.Transaction, activatorID uint64) (entity.PartialPercentageActivator, *errs.Error)
	FindPercentageActivatorByID(ct context.Context, activatorID uint64) (entity.PartialPercentageActivator, *errs.Error)
	CreatePercentageActivator(ct context.Context, tx *transaction.Transaction, activatorID uint64, activator entity.PartialPercentageActivator) *errs.Error
	UpdatePercentageActivator(ct context.Context, tx *transaction.Transaction, activatorID uint64, activator entity.PartialPercentageActivator) *errs.Error
	DeletePercentageActivator(ct context.Context, tx *transaction.Transaction, activatorID uint64) *errs.Error
}
