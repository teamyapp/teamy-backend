package daov2

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskLink interface {
	FindTaskLinkByID(ct context.Context, tx *transaction.Transaction, taskLinkID uint64) (entity.TaskLink, *errs.Error)
	FindLinksByTaskID(ct context.Context, tx *transaction.Transaction, taskID uint64) ([]entity.TaskLink, *errs.Error)
	CreateTaskLink(ct context.Context, tx *transaction.Transaction, taskLinkEntity entity.TaskLink) *errs.Error
	DeleteTaskLink(ct context.Context, tx *transaction.Transaction, taskLinkID uint64) *errs.Error
}
