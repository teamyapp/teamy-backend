package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type ActivatorTypeRelation interface {
	FindActivatorTypeByIDWithTx(ct context.Context, tx *transaction.Transaction, activatorID uint64) (entity.ActivatorType, *errs.Error)
	FindActivatorTypeByID(ct context.Context, activatorID uint64) (entity.ActivatorType, *errs.Error)
	CreateActivatorTypeRelation(ct context.Context, tx *transaction.Transaction, activatorTypeRelation entity.ActivatorTypeRelation) *errs.Error
}
