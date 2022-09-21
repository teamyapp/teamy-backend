package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/core/entity"
)

type Task interface {
	FindTaskByID(ct context.Context, taskID uint64) (entity.Task, error)
	FindTaskByCommentsThreadID(ct context.Context, commentThreadID uint64) (entity.Task, error)
	FindTasksByIDs(ct context.Context, taskIDs []uint64) ([]entity.Task, error)
	FindTasksByTeamID(ct context.Context, teamID uint64) ([]entity.Task, error)
	FindAllTasks(ct context.Context) ([]entity.Task, error)
	CreateTask(ct context.Context, task entity.Task) error
	UpdateTask(ct context.Context, task entity.Task) error
	DeleteTask(ct context.Context, taskID uint64) error
}
