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

const attachmentFileUploadSessionDaoName = "AttachmentFileUploadSession"

type AttachmentFileUploadSession struct {
	metrics dao.Metrics
}

var _ dao.AttachmentFileUploadSession = (*AttachmentFileUploadSession)(nil)

func (t AttachmentFileUploadSession) FindAttachmentFileUploadSessionWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	attachmentListID uint64,
	fileUploadSessionID uint64,
) (entity.AttachmentFileUploadSession, *errs.Error) {
	t.metrics.ReportDaoOperation(attachmentFileUploadSessionDaoName, "FindAttachmentFileUploadSessionWithTx")
	attachmentFileUploadSession := entity.AttachmentFileUploadSession{}
	err := tx.SQLTx().QueryRow(`
		SELECT
			attachment_list_id,
			file_upload_session_id,
			is_completed,
			created_at,
			updated_at
		FROM attachment_file_upload_session
		WHERE attachment_list_id = $1 AND file_upload_session_id = $2;`,
		attachmentListID, fileUploadSessionID).
		Scan(
			&attachmentFileUploadSession.AttachmentListID,
			&attachmentFileUploadSession.FileUploadSessionID,
			&attachmentFileUploadSession.IsCompleted,
			&attachmentFileUploadSession.CreatedAt,
			&attachmentFileUploadSession.UpdatedAt,
		)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.AttachmentFileUploadSession{}, errs.NewError(errs.NotFound, fmt.Sprintf(
				"AttachmentFileUploadSession not found: attachmentListID=%v, fileUploadSessionID=%v", attachmentListID, fileUploadSessionID))
		}

		return entity.AttachmentFileUploadSession{}, errs.NewError(errs.Unknown, err.Error())
	}

	return attachmentFileUploadSession, nil
}

func (t AttachmentFileUploadSession) CreateAttachmentFileUploadSession(
	ct context.Context,
	tx *transaction.Transaction,
	attachmentFileUploadSession entity.AttachmentFileUploadSession,
) *errs.Error {
	t.metrics.ReportDaoOperation(attachmentFileUploadSessionDaoName, "CreateAttachmentFileUploadSession")
	_, err := tx.SQLTx().Exec(`
		INSERT INTO attachment_file_upload_session
		(
			attachment_list_id,
			file_upload_session_id,
			is_completed,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5);`,
		attachmentFileUploadSession.AttachmentListID,
		attachmentFileUploadSession.FileUploadSessionID,
		attachmentFileUploadSession.IsCompleted,
		attachmentFileUploadSession.CreatedAt,
		attachmentFileUploadSession.UpdatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t AttachmentFileUploadSession) UpdateAttachmentFileUploadSession(
	ct context.Context,
	tx *transaction.Transaction,
	attachmentFileUploadSession entity.AttachmentFileUploadSession,
) *errs.Error {
	t.metrics.ReportDaoOperation(attachmentFileUploadSessionDaoName, "UpdateAttachmentFileUploadSession")
	_, err := tx.SQLTx().Exec(`
		UPDATE attachment_file_upload_session
		SET
			is_completed = $1,
			created_at= $2,
			updated_at = $3
		WHERE attachment_id = $4 AND file_upload_session_id = $5;`,
		attachmentFileUploadSession.AttachmentListID,
		attachmentFileUploadSession.FileUploadSessionID,
		attachmentFileUploadSession.IsCompleted,
		attachmentFileUploadSession.CreatedAt,
		attachmentFileUploadSession.UpdatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewAttachmentFileUploadSession(metrics dao.Metrics) *AttachmentFileUploadSession {
	return &AttachmentFileUploadSession{
		metrics: metrics,
	}
}
