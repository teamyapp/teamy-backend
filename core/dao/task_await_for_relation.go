package dao

import (
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskAwaitForRelation interface {
	FindAwaitingTaskIDs(waitForTaskID uint64) ([]uint64, error)
	FindAwaitForTaskIDs(waitingTaskID uint64) ([]uint64, error)
	CreateRelation(relation entity.TaskAwaitForRelation) error
	DeleteRelation(waitingTaskID uint64, awaitForTaskID uint64) error
}
