package daotest

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskAwaitForRelation struct {
	db *dbtest.InMemoryDB
}

var _ dao.TaskAwaitForRelation = (*TaskAwaitForRelation)(nil)

func (t TaskAwaitForRelation) FindAwaitingTaskIDs(ct context.Context, waitForTaskID uint64) ([]uint64, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t TaskAwaitForRelation) FindAwaitForTaskIDs(ct context.Context, waitingTaskID uint64) ([]uint64, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t TaskAwaitForRelation) CreateRelation(ct context.Context, relation entity.TaskAwaitForRelation) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (t TaskAwaitForRelation) DeleteRelation(ct context.Context, waitingTaskID uint64, awaitForTaskID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewTaskAwaitForRelation(db *dbtest.InMemoryDB) TaskAwaitForRelation {
	return TaskAwaitForRelation{db: db}
}
