package daov2

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Task interface {
	FindTaskByID(ct context.Context, tx *transaction.Transaction, taskID uint64) (entity.Task, *errs.Error)
	FindTaskByCommentsThreadID(ct context.Context, tx *transaction.Transaction, commentThreadID uint64) (entity.Task, *errs.Error)
	FindTasksByIDs(ct context.Context, tx *transaction.Transaction, taskIDs []uint64) ([]entity.Task, *errs.Error)
	FindTasksByTeamID(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.Task, *errs.Error)
	FindAllTasks(ct context.Context, tx *transaction.Transaction) ([]entity.Task, *errs.Error)
	CreateTask(ct context.Context, tx *transaction.Transaction, task entity.Task) *errs.Error
	UpdateTask(ct context.Context, tx *transaction.Transaction, task entity.Task) *errs.Error
	DeleteTask(ct context.Context, tx *transaction.Transaction, taskID uint64) *errs.Error
}
