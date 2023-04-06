package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type UserFileUploadSession struct {
	logger telemetry.Logger
	db     *sql.DB
}

var _ dao.UserFileUploadSession = (*UserFileUploadSession)(nil)

func (u UserFileUploadSession) FindUserFileUploadSessionByUserID(
	ct context.Context,
	userID uint64,
	userFileUploadSessionType entity.UserFileUploadSessionType,
	fileUploadSessionID uint64,
) (entity.UserFileUploadSession, *errs.Error) {
	userFileUploadSession := entity.UserFileUploadSession{}
	err := u.db.QueryRow(`
		SELECT
			user_id,
			type,
			file_upload_session_id,
			is_completed,
			created_at,
			updated_at
		FROM user_file_upload_session
		WHERE user_id = $1 AND type = $2 AND file_upload_session_id = $3;`,
		userID, userFileUploadSessionType, fileUploadSessionID).
		Scan(
			&userFileUploadSession.UserID,
			&userFileUploadSession.Type,
			&userFileUploadSession.FileUploadSessionID,
			&userFileUploadSession.IsCompleted,
			&userFileUploadSession.CreatedAt,
			&userFileUploadSession.UpdatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code: errs.NotFound,
			Message: fmt.Sprintf("UserFileUploadSession not found: userID=%v, userFileUploadSessionType=%v",
				userID,
				userFileUploadSessionType),
		}
		u.logger.ErrorWithContext(ct, internalErr)
		return entity.UserFileUploadSession{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		u.logger.ErrorWithContext(ct, internalErr)
		return entity.UserFileUploadSession{}, internalErr
	}

	return userFileUploadSession, nil
}

func (u UserFileUploadSession) CreateUserFileUploadSession(
	ct context.Context,
	userFileUploadSession entity.UserFileUploadSession,
) *errs.Error {
	_, err := u.db.Exec(`
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		u.logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (u UserFileUploadSession) UpdateUserFileUploadSession(
	ct context.Context,
	userFileUploadSession entity.UserFileUploadSession,
) *errs.Error {
	_, err := u.db.Exec(`
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		u.logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func NewUserFileUploadSession(logger telemetry.Logger, sqlDB *sql.DB) UserFileUploadSession {
	return UserFileUploadSession{logger: logger, db: sqlDB}
}
