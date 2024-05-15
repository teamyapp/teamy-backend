package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const taskFileUploadSessionDaoName = "TaskFileUploadSession"

type TaskFileUploadSession struct {
	metrics dao.Metrics
}

var _ dao.TaskFileUploadSession = (*TaskFileUploadSession)(nil)

func (t TaskFileUploadSession) FindTaskFileUploadSessionByTaskIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	taskID uint64,
	taskFileUploadSessionType entity.TaskFileUploadSessionType,
	fileUploadSessionID uint64,
) (entity.TaskFileUploadSession, *errs.Error) {
	t.metrics.ReportDaoOperation(taskFileUploadSessionDaoName, "FindTaskFileUploadSessionByTaskIDWithTx")
	taskFileUploadSession := entity.TaskFileUploadSession{}
	err := tx.SQLTx().QueryRow(`
		SELECT
			task_id,
			type,
			file_upload_session_id,
			is_completed,
			created_at,
			updated_at
		FROM task_file_upload_session
		WHERE task_id = $1 AND type = $2 AND file_upload_session_id = $3;`,
		taskID, taskFileUploadSessionType, fileUploadSessionID).
		Scan(
			&taskFileUploadSession.TaskID,
			&taskFileUploadSession.Type,
			&taskFileUploadSession.FileUploadSessionID,
			&taskFileUploadSession.IsCompleted,
			&taskFileUploadSession.CreatedAt,
			&taskFileUploadSession.UpdatedAt,
		)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.TaskFileUploadSession{}, errs.NewError(errs.NotFound, fmt.Sprintf(
				"TaskFileUploadSession not found: taskID=%v, taskFileUploadSessionType=%v", taskID, taskFileUploadSessionType))
		}

		return entity.TaskFileUploadSession{}, errs.NewError(errs.Unknown, err.Error())
	}

	return taskFileUploadSession, nil
}

func (t TaskFileUploadSession) CreateTaskFileUploadSession(
	ct context.Context,
	tx *transaction.Transaction,
	taskFileUploadSession entity.TaskFileUploadSession,
) *errs.Error {
	t.metrics.ReportDaoOperation(taskFileUploadSessionDaoName, "CreateTaskFileUploadSession")
	_, err := tx.SQLTx().Exec(`
		INSERT INTO task_file_upload_session
		(
			task_id,
			type,
			file_upload_session_id,
			is_completed,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6);`,
		taskFileUploadSession.TaskID,
		taskFileUploadSession.Type,
		taskFileUploadSession.FileUploadSessionID,
		taskFileUploadSession.IsCompleted,
		taskFileUploadSession.CreatedAt,
		taskFileUploadSession.UpdatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t TaskFileUploadSession) UpdateTaskFileUploadSession(
	ct context.Context,
	tx *transaction.Transaction,
	taskFileUploadSession entity.TaskFileUploadSession,
) *errs.Error {
	t.metrics.ReportDaoOperation(taskFileUploadSessionDaoName, "UpdateTaskFileUploadSession")
	_, err := tx.SQLTx().Exec(`
		UPDATE task_file_upload_session
		SET
			task_id = $1,
			type = $2,
			file_upload_session_id = $3,
			is_completed = $4,
			created_at= $5,
			updated_at = $6
		WHERE task_id = $7 AND type = $8 AND file_upload_session_id = $9;`,
		taskFileUploadSession.TaskID,
		taskFileUploadSession.Type,
		taskFileUploadSession.FileUploadSessionID,
		taskFileUploadSession.IsCompleted,
		taskFileUploadSession.CreatedAt,
		taskFileUploadSession.UpdatedAt,
		taskFileUploadSession.TaskID,
		taskFileUploadSession.Type,
		taskFileUploadSession.FileUploadSessionID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTaskFileUploadSession(metrics dao.Metrics) *TaskFileUploadSession {
	return &TaskFileUploadSession{
		metrics: metrics,
	}
}
