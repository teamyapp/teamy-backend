package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Task interface {
	FindTaskByID(ct context.Context, taskID uint64) (entity.Task, *errs.Error)
	FindTaskByCommentsThreadID(ct context.Context, commentThreadID uint64) (entity.Task, *errs.Error)
	FindTasksByIDs(ct context.Context, taskIDs []uint64) ([]entity.Task, *errs.Error)
	FindTasksByTeamID(ct context.Context, teamID uint64) ([]entity.Task, *errs.Error)
	FindAllTasks(ct context.Context) ([]entity.Task, *errs.Error)
	CreateTask(ct context.Context, task entity.Task) *errs.Error
	UpdateTask(ct context.Context, task entity.Task) *errs.Error
	DeleteTask(ct context.Context, taskID uint64) *errs.Error
}
