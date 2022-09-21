package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintTaskRelation interface {
	FindTaskIDsBySprintID(ct context.Context, sprintID uint64) ([]uint64, error)
	FindSprintIDsByTaskID(ct context.Context, taskID uint64) ([]uint64, error)
	CreateSprintTaskRelation(ct context.Context, relation entity.SprintTaskRelation) error
	DeleteSprintTaskRelation(ct context.Context, sprintID uint64, taskID uint64) error
}
