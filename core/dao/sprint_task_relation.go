package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintTaskRelation interface {
	FindTaskIDsBySprintID(ct context.Context, sprintID uint64) ([]uint64, *errs.Error)
	FindSprintIDsByTaskID(ct context.Context, taskID uint64) ([]uint64, *errs.Error)
	CreateSprintTaskRelation(ct context.Context, relation entity.SprintTaskRelation) *errs.Error
	DeleteSprintTaskRelation(ct context.Context, sprintID uint64, taskID uint64) *errs.Error
}
