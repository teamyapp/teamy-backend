package dao

import (
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintTaskRelation interface {
	FindTaskIDsBySprintID(sprintID uint64) ([]uint64, error)
	FindSprintIDsByTaskID(taskID uint64) ([]uint64, error)
	CreateSprintTaskRelation(relation entity.SprintTaskRelation) error
	DeleteSprintTaskRelation(sprintID uint64, taskID uint64) error
}
