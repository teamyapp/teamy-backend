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

var _ dao.AppVersion = (*AppVersion)(nil)

type AppVersion struct {
	logger             telemetry.Logger
	transactionFactory transaction.Factory
}

func (a AppVersion) FindAppVersionsByAppID(ct context.Context, appID uint64) ([]entity.AppVersion, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return a.FindAppVersionsByAppIDWithTx(ct, tx, appID)
}

func (a AppVersion) FindAppVersionByAppIDAndVersionNumber(ct context.Context, appID uint64,
	versionNumber int32) (entity.AppVersion, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.AppVersion{}, err
	}

	defer tx.Rollback()
	return a.FindAppVersionByAppIDAndVersionNumberWithTx(ct, tx, appID, versionNumber)
}

func (a AppVersion) FindAppVersionByAppIDAndVersionNumberWithTx(ct context.Context, tx *transaction.Transaction,
	appID uint64, versionNumber int32) (entity.AppVersion, *errs.Error) {
	appVersion := entity.AppVersion{}
	err := tx.SQLTx().QueryRow(`
	SELECT
	    app_id,
	    version_number,
	    icon_url,
	    has_ui_extension,
	    ui_extension_entrypoint_path,
	    is_public,
	    created_at,
	    updated_at,
	    changes
	FROM app_version
	WHERE app_id = $1 AND version_number = $2;
`,
		appID,
		versionNumber).
		Scan(
			&appVersion.AppID,
			&appVersion.VersionNumber,
			&appVersion.IconURL,
			&appVersion.HasUIExtension,
			&appVersion.UIExtensionEntrypointPath,
			&appVersion.IsPublic,
			&appVersion.CreatedAt,
			&appVersion.UpdateAt,
			&appVersion.Changes,
		)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.AppVersion{}, errs.NewError(errs.NotFound,
				fmt.Sprintf("app version not found: appID=%v, versionNum=%v", appID, versionNumber))
		}

		return entity.AppVersion{}, errs.NewError(errs.Unknown, err.Error())
	}

	return appVersion, nil
}

func (a AppVersion) FindAppVersionsByAppIDWithTx(ct context.Context, tx *transaction.Transaction,
	appID uint64) ([]entity.AppVersion, *errs.Error) {
	rows, err := tx.SQLTx().Query(`
	SELECT
	    app_id,
	    version_number,
	    icon_url,
	    has_ui_extension,
	    ui_extension_entrypoint_path,
	    is_public,
	    created_at,
	    updated_at,
	    changes
	FROM app_version
	WHERE app_id = $1;
`,
		appID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	appVersions := make([]entity.AppVersion, 0)
	for rows.Next() {
		appVersion := entity.AppVersion{}
		err = rows.Scan(
			&appVersion.AppID,
			&appVersion.VersionNumber,
			&appVersion.IconURL,
			&appVersion.HasUIExtension,
			&appVersion.UIExtensionEntrypointPath,
			&appVersion.IsPublic,
			&appVersion.CreatedAt,
			&appVersion.UpdateAt,
			&appVersion.Changes,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		appVersions = append(appVersions, appVersion)
	}

	return appVersions, nil
}

func (a AppVersion) CreateAppVersion(ct context.Context, tx *transaction.Transaction,
	appVersion entity.AppVersion) *errs.Error {
	// The uniqueness of version number is guaranteed by the primary key constraint of app_version table.
	// if try to insert into a new app version with the same version number, one should fail since db doesn't allow duplicate primary key
	// client could see according error and should retry
	_, err := tx.SQLTx().Exec(`
	INSERT INTO app_version
	(
	 	app_id,
	 	version_number,
	 	icon_url,
	 	has_ui_extension,
	 	ui_extension_entrypoint_path,
	 	is_public,
	 	created_at,
	 	updated_at,
	 	changes
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`,
		appVersion.AppID,
		appVersion.VersionNumber,
		appVersion.IconURL,
		appVersion.HasUIExtension,
		appVersion.UIExtensionEntrypointPath,
		appVersion.IsPublic,
		appVersion.CreatedAt,
		appVersion.UpdateAt,
		appVersion.Changes,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a AppVersion) UpdateAppVersion(ct context.Context, tx *transaction.Transaction,
	appVersion entity.AppVersion) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		UPDATE app_version
		SET
			icon_url = $1,
			has_ui_extension = $2,
			ui_extension_entrypoint_path = $3,
			is_public = $4,
			changes = $5,
			updated_at = $6
		WHERE app_id = $7 AND version_number = $8;`,
		appVersion.IconURL,
		appVersion.HasUIExtension,
		appVersion.UIExtensionEntrypointPath,
		appVersion.IsPublic,
		appVersion.Changes,
		appVersion.UpdateAt,
		appVersion.AppID,
		appVersion.VersionNumber,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a AppVersion) FindMaxVersionNumberWithTx(ct context.Context, tx *transaction.Transaction,
	appID uint64) (int32, *errs.Error) {
	var maxVersion int32
	row := tx.SQLTx().QueryRow(`
	SELECT MAX(version_number)
	FROM app_version
	WHERE app_id = $1
`,
		appID,
	)

	err := row.Err()

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errs.NewError(errs.NotFound, fmt.Sprintf("max VersionNumber not found: appID=%v", appID))
		}

		return 0, errs.NewError(errs.Unknown, err.Error())
	}

	err = row.Scan(&maxVersion)
	if err != nil {
		return 0, errs.NewError(errs.Unknown, err.Error())
	}

	return maxVersion, nil
}

func (a AppVersion) DeleteAppVersion(ct context.Context, tx *transaction.Transaction, appID uint64,
	versionNumber int32) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM app_version
		WHERE app_id = $1
		AND version_number = $2;
		`,
		appID,
		versionNumber)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewAppVersion(logger telemetry.Logger, sqlDB *sql.DB) AppVersion {
	return AppVersion{logger: logger}
}
