package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppVersion struct {
	logger             telemetry.Logger
	transactionFactory transaction.Factory
}

var _ dao.AppVersion = (*AppVersion)(nil)

func (a *AppVersion) FindAppVersionsByAppIDWithTx(ct context.Context, tx *transaction.Transaction, appID uint64) ([]entity.AppVersion, *errs.Error) {
	appVersions := []entity.AppVersion{}
	rows, err := tx.SQLTx().Query(`
		SELECT
			app_id,
			number,
			app_name,
			has_ui_extension,
			description,
			created_at,
			created_by_user_id,
			is_ready
		FROM app_version
		WHERE app_id = $1;`,
		appID,
	)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		appVersion := entity.AppVersion{}
		err := rows.Scan(
			&appVersion.AppID,
			&appVersion.Number,
			&appVersion.AppName,
			&appVersion.HasUiExtension,
			&appVersion.Description,
			&appVersion.CreatedAt,
			&appVersion.CreatedByUserID,
			&appVersion.IsReady,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		appVersions = append(appVersions, appVersion)
	}

	return appVersions, nil
}

func (a *AppVersion) FindAppVersionsByAppID(ct context.Context, appID uint64) ([]entity.AppVersion, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}

	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	appVersions, err := a.FindAppVersionsByAppIDWithTx(ct, tx, appID)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return appVersions, nil
}

func (a *AppVersion) FindMaxVersionNumberWithTx(ct context.Context, tx *transaction.Transaction, appID uint64) (int, *errs.Error) {
	var versionNumber int
	err := tx.SQLTx().QueryRow(`
		SELECT
			MAX(number)
		FROM app_version
		WHERE app_id = $1;`,
		appID).
		Scan(
			&versionNumber,
		)
	if err != nil {
		return 0, errs.NewError(errs.Unknown, err.Error())
	}

	return versionNumber, nil
}

func (a *AppVersion) CreateAppVersion(ct context.Context, tx *transaction.Transaction, appVersion entity.AppVersion) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		INSERT INTO app_version (
			app_id,
			number,
			app_name,
			has_ui_extension,
			description,
			created_at,
			created_by_user_id,
			is_ready
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7,
			$8
		);`,
		appVersion.AppID,
		appVersion.Number,
		appVersion.AppName,
		appVersion.HasUiExtension,
		appVersion.Description,
		appVersion.CreatedAt,
		appVersion.CreatedByUserID,
		appVersion.IsReady,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a *AppVersion) DeleteAppVersion(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM app_version
		WHERE app_id = $1 AND number = $2;`,
		appID,
		versionNumber,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a *AppVersion) FindAppVersionByAppIDAndVersionNumberWithTx(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int) (entity.AppVersion, *errs.Error) {
	appVersion := entity.AppVersion{}
	err := tx.SQLTx().QueryRow(`
		SELECT
			app_id,
			number,
			app_name,
			has_ui_extension,
			description,
			created_at,
			created_by_user_id,
			is_ready
		FROM app_version
		WHERE app_id = $1 AND number = $2;`,
		appID,
		versionNumber,
	).Scan(
		&appVersion.AppID,
		&appVersion.Number,
		&appVersion.AppName,
		&appVersion.HasUiExtension,
		&appVersion.Description,
		&appVersion.CreatedAt,
		&appVersion.CreatedByUserID,
		&appVersion.IsReady,
	)
	if err != nil {
		return entity.AppVersion{}, errs.NewError(errs.Unknown, err.Error())
	}

	return appVersion, nil
}

func NewAppVersion(logger telemetry.Logger, transactionFactory transaction.Factory) *AppVersion {
	return &AppVersion{
		logger:             logger,
		transactionFactory: transactionFactory,
	}
}
