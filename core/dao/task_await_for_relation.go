package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskAwaitForRelation interface {
	FindAwaitingTaskIDs(ct context.Context, waitForTaskID uint64) ([]uint64, *errs.Error)
	FindAwaitForTaskIDs(ct context.Context, waitingTaskID uint64) ([]uint64, *errs.Error)
	CreateRelation(ct context.Context, relation entity.TaskAwaitForRelation) *errs.Error
	DeleteRelation(ct context.Context, waitingTaskID uint64, awaitForTaskID uint64) *errs.Error
}
