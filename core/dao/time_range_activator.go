package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao/entity"
)

type TimeRangeActivator interface {
	FindTimeRangeActivatorByIDWithTx(ct context.Context, tx *transaction.Transaction, activatorID uint64) (entity.PartialTimeRangeActivator, *errs.Error)
	FindTimeRangeActivatorByID(ct context.Context, activatorID uint64) (entity.PartialTimeRangeActivator, *errs.Error)
	CreateTimeRangeActivator(ct context.Context, tx *transaction.Transaction, activatorID uint64, activator entity.PartialTimeRangeActivator) *errs.Error
	UpdateTimeRangeActivator(ct context.Context, tx *transaction.Transaction, activatorID uint64, activator entity.PartialTimeRangeActivator) *errs.Error
	DeleteTimeRangeActivator(ct context.Context, tx *transaction.Transaction, activatorID uint64) *errs.Error
}
