package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskAwaitForRelation interface {
	FindAwaitingTaskIDsWithTx(ct context.Context, tx *transaction.Transaction, waitForTaskID uint64) ([]uint64, *errs.Error)
	FindAwaitForTaskIDsWithTx(ct context.Context, tx *transaction.Transaction, waitingTaskID uint64) ([]uint64, *errs.Error)
	CreateRelation(ct context.Context, tx *transaction.Transaction, relation entity.TaskAwaitForRelation) *errs.Error
	DeleteRelation(ct context.Context, tx *transaction.Transaction, waitingTaskID uint64, awaitForTaskID uint64) *errs.Error
}
