package daov2

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Task interface {
	FindTaskByID(ct context.Context, tx *sql.Tx, taskID uint64) (entity.Task, *errs.Error)
	FindTaskByCommentsThreadID(ct context.Context, tx *sql.Tx, commentThreadID uint64) (entity.Task, *errs.Error)
	FindTasksByIDs(ct context.Context, tx *sql.Tx, taskIDs []uint64) ([]entity.Task, *errs.Error)
	FindTasksByTeamID(ct context.Context, tx *sql.Tx, teamID uint64) ([]entity.Task, *errs.Error)
	FindAllTasks(ct context.Context, tx *sql.Tx) ([]entity.Task, *errs.Error)
	CreateTask(ct context.Context, tx *sql.Tx, task entity.Task) *errs.Error
	UpdateTask(ct context.Context, tx *sql.Tx, task entity.Task) *errs.Error
	DeleteTask(ct context.Context, tx *sql.Tx, taskID uint64) *errs.Error
}
