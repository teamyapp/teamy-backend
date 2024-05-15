package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskFileUploadSession interface {
	FindTaskFileUploadSessionByTaskIDWithTx(
		ct context.Context,
		tx *transaction.Transaction,
		taskID uint64,
		taskFileUploadSessionType entity.TaskFileUploadSessionType,
		fileUploadSessionID uint64,
	) (entity.TaskFileUploadSession, *errs.Error)
	CreateTaskFileUploadSession(ct context.Context, tx *transaction.Transaction, taskFileUploadSession entity.TaskFileUploadSession) *errs.Error
	UpdateTaskFileUploadSession(ct context.Context, tx *transaction.Transaction, taskFileUploadSession entity.TaskFileUploadSession) *errs.Error
}
