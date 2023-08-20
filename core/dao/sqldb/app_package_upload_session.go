package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppPackageUploadSession struct {
	logger             telemetry.Logger
	transactionFactory transaction.Factory
}

var _ dao.AppPackageUploadSession = (*AppPackageUploadSession)(nil)

func (a *AppPackageUploadSession) FindAppPackageUploadSessionWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	appID uint64,
	userID uint64,
	versionNumber int32,
	fileUploadSessionID uint64,
) (entity.AppPackageUploadSession, *errs.Error) {
	appPackageUploadSession := entity.AppPackageUploadSession{}
	err := tx.SQLTx().QueryRow(`
		SELECT
			app_id,
			user_id,
			file_upload_session_id,
			version_number,
			is_completed,
			created_at,
			updated_at
		FROM app_package_upload_session
		WHERE app_id = $1 
		    AND user_id = $2 
		    AND version_number = $3 
		    AND file_upload_session_id = $4;`,
		appID,
		userID,
		versionNumber,
		fileUploadSessionID).
		Scan(
			&appPackageUploadSession.AppID,
			&appPackageUploadSession.UserID,
			&appPackageUploadSession.FileUploadSessionID,
			&appPackageUploadSession.VersionNumber,
			&appPackageUploadSession.IsCompleted,
			&appPackageUploadSession.CreatedAt,
			&appPackageUploadSession.UpdatedAt,
		)
	if errors.Is(err, sql.ErrNoRows) {
		return entity.AppPackageUploadSession{}, errs.NewError(errs.NotFound, fmt.Sprintf("AppPackageUploadSession not found: appID=%v, userID=%v, versionNumber=%v",
			appID,
			userID,
			versionNumber))
	}

	if err != nil {
		return entity.AppPackageUploadSession{}, errs.NewError(errs.Unknown, err.Error())
	}

	return appPackageUploadSession, nil
}

func (a *AppPackageUploadSession) CreateAppPackageUploadSession(ct context.Context, tx *transaction.Transaction, AppPackageUploadSession entity.AppPackageUploadSession) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		INSERT INTO app_package_upload_session
		(
			app_id,
			user_id,
			file_upload_session_id,
			version_number,
			is_completed,
			created_at,
			updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7);`,
		AppPackageUploadSession.AppID,
		AppPackageUploadSession.UserID,
		AppPackageUploadSession.FileUploadSessionID,
		AppPackageUploadSession.VersionNumber,
		AppPackageUploadSession.IsCompleted,
		AppPackageUploadSession.CreatedAt,
		AppPackageUploadSession.UpdatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a *AppPackageUploadSession) UpdateAppPackageFileUploadSession(
	ct context.Context,
	tx *transaction.Transaction,
	session entity.AppPackageUploadSession,
) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		UPDATE app_package_upload_session
		SET
			is_completed = $1,
			updated_at = $2
		WHERE app_id = $3 AND user_id = $4 AND version_number = $5 and file_upload_session_id = $6;`,
		AppPackageUploadSession.IsCompleted,
		AppPackageUploadSession.UpdatedAt,
		AppPackageUploadSession.AppID,
		AppPackageUploadSession.UserID,
		AppPackageUploadSession.VersionNumber,
		AppPackageUploadSession.FileUploadSessionID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewAppPackageUploadSession(logger telemetry.Logger, transactionFactory transaction.Factory) *AppPackageUploadSession {
	return &AppPackageUploadSession{
		logger:             logger,
		transactionFactory: transactionFactory,
	}
}
