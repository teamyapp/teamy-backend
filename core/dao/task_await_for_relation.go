package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskAwaitForRelation interface {
	FindAwaitingTaskIDs(ct context.Context, waitForTaskID uint64) ([]uint64, error)
	FindAwaitForTaskIDs(ct context.Context, waitingTaskID uint64) ([]uint64, error)
	CreateRelation(ct context.Context, relation entity.TaskAwaitForRelation) error
	DeleteRelation(ct context.Context, waitingTaskID uint64, awaitForTaskID uint64) error
}
