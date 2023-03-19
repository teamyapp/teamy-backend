package daotestv2

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskAwaitForRelation struct {
	db *dbtest.InMemoryDB
}

var _ daov2.TaskAwaitForRelation = (*TaskAwaitForRelation)(nil)

func (t TaskAwaitForRelation) FindAwaitingTaskIDsWithTx(ct context.Context, tx *transaction.Transaction, waitForTaskID uint64) ([]uint64, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t TaskAwaitForRelation) FindAwaitForTaskIDsWithTx(ct context.Context, tx *transaction.Transaction, waitingTaskID uint64) ([]uint64, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t TaskAwaitForRelation) CreateRelation(ct context.Context, tx *transaction.Transaction, relation entity.TaskAwaitForRelation) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (t TaskAwaitForRelation) DeleteRelation(ct context.Context, tx *transaction.Transaction, waitingTaskID uint64, awaitForTaskID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewTaskAwaitForRelation(db *dbtest.InMemoryDB) TaskAwaitForRelation {
	return TaskAwaitForRelation{db: db}
}
