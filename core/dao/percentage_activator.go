package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type PercentageActivator interface {
	FindPercentageActivatorByID(ct context.Context, ActivatorID uint64) (entity.PercentageActivator, *errs.Error)
	CreatePercentageActivator(ct context.Context, tx *transaction.Transaction, activator entity.PercentageActivator) *errs.Error
}
