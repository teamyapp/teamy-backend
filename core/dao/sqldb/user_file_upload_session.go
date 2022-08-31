package sqldb

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type UserFileUploadSession struct {
	dataCollector obs.DataCollector
	db            *sql.DB
}

var _ dao.UserFileUploadSession = (*UserFileUploadSession)(nil)

func (u UserFileUploadSession) FindUserFileUploadSessionByUserID(
	userID uint64,
	userFileUploadSessionType entity.UserFileUploadSessionType,
	fileUploadSessionID uint64,
) (entity.UserFileUploadSession, error) {
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
		return entity.UserFileUploadSession{}, dao.ErrNotFound(fmt.Sprintf(
			"UserFileUploadSession not found: userID=%v, type=%v",
			userID,
			userFileUploadSessionType))
	}

	if err != nil {
		u.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return userFileUploadSession, err
}

func (u UserFileUploadSession) CreateUserFileUploadSession(userFileUploadSession entity.UserFileUploadSession) error {
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
		u.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (u UserFileUploadSession) UpdateUserFileUploadSession(userFileUploadSession entity.UserFileUploadSession) error {
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
		u.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewUserFileUploadSession(dataCollector obs.DataCollector, sqlDB *sql.DB) UserFileUploadSession {
	return UserFileUploadSession{dataCollector: dataCollector, db: sqlDB}
}
