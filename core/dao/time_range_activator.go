package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TimeRangeActivator interface {
	FindTimeRangeActivatorByID(ct context.Context, activatorID uint64) (entity.TimeRangeActivator, *errs.Error)
	CreateTimeRangeActivator(ct context.Context, tx *transaction.Transaction, activator entity.TimeRangeActivator) *errs.Error
}
