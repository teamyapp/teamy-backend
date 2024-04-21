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

const userFileUploadSessionDaoName = "UserFileUploadSession"

type UserFileUploadSession struct {
	metrics dao.Metrics
}

var _ dao.UserFileUploadSession = (*UserFileUploadSession)(nil)

func (u UserFileUploadSession) FindUserFileUploadSessionByUserIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	userID uint64,
	userFileUploadSessionType entity.UserFileUploadSessionType,
	fileUploadSessionID uint64,
) (entity.UserFileUploadSession, *errs.Error) {
	u.metrics.ReportDaoOperation(userFileUploadSessionDaoName, "FindUserFileUploadSessionByUserIDWithTx")
	userFileUploadSession := entity.UserFileUploadSession{}
	err := tx.SQLTx().QueryRow(`
		SELECT
			user_id,
			type,
			file_upload_session_id,
			is_completed,
			created_at,
			updated_at
		FROM user_file_upload_session
		WHERE user_id = $1 AND type = $2 AND file_upload_session_id = $3;`,
		userID,
		userFileUploadSessionType,
		fileUploadSessionID).
		Scan(
			&userFileUploadSession.UserID,
			&userFileUploadSession.Type,
			&userFileUploadSession.FileUploadSessionID,
			&userFileUploadSession.IsCompleted,
			&userFileUploadSession.CreatedAt,
			&userFileUploadSession.UpdatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.UserFileUploadSession{}, errs.NewError(errs.NotFound, fmt.Sprintf("UserFileUploadSession not found: userID=%v, userFileUploadSessionType=%v",
			userID,
			userFileUploadSessionType))
	}

	if err != nil {
		return entity.UserFileUploadSession{}, errs.NewError(errs.Unknown, err.Error())
	}

	return userFileUploadSession, nil
}

func (u UserFileUploadSession) CreateUserFileUploadSession(
	ct context.Context,
	tx *transaction.Transaction,
	userFileUploadSession entity.UserFileUploadSession,
) *errs.Error {
	u.metrics.ReportDaoOperation(userFileUploadSessionDaoName, "CreateUserFileUploadSession")
	_, err := tx.SQLTx().Exec(`
		INSERT INTO user_file_upload_session
		(
			user_id,
			type,
			file_upload_session_id,
			is_completed,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6);`,
		userFileUploadSession.UserID,
		userFileUploadSession.Type,
		userFileUploadSession.FileUploadSessionID,
		userFileUploadSession.IsCompleted,
		userFileUploadSession.CreatedAt,
		userFileUploadSession.UpdatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (u UserFileUploadSession) UpdateUserFileUploadSession(
	ct context.Context,
	tx *transaction.Transaction,
	userFileUploadSession entity.UserFileUploadSession,
) *errs.Error {
	u.metrics.ReportDaoOperation(userFileUploadSessionDaoName, "UpdateUserFileUploadSession")
	_, err := tx.SQLTx().Exec(`
		UPDATE user_file_upload_session
		SET
			user_id = $1,
			type = $2,
			file_upload_session_id = $3,
			is_completed = $4,
			created_at= $5,
			updated_at = $6
		WHERE user_id = $7 AND type = $8 AND file_upload_session_id = $9;`,
		userFileUploadSession.UserID,
		userFileUploadSession.Type,
		userFileUploadSession.FileUploadSessionID,
		userFileUploadSession.IsCompleted,
		userFileUploadSession.CreatedAt,
		userFileUploadSession.UpdatedAt,
		userFileUploadSession.UserID,
		userFileUploadSession.Type,
		userFileUploadSession.FileUploadSessionID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewUserFileUploadSession(metrics dao.Metrics) UserFileUploadSession {
	return UserFileUploadSession{
		metrics: metrics,
	}
}
