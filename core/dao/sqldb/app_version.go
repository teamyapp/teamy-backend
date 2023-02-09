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

var _ dao.AppVersion = (*AppVersion)(nil)

type AppVersion struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

func (a AppVersion) FindAppVersionByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int32) (entity.AppVersion, *errs.Error) {
	appVersion := entity.AppVersion{}
	err := a.db.QueryRow(`
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

	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("app version not found: appID=%v, versionNum=%v", appID, versionNumber),
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.AppVersion{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.AppVersion{}, internalErr
	}

	return appVersion, nil
}

func (a AppVersion) FindAppVersionsByAppID(ct context.Context, appID uint64) ([]entity.AppVersion, *errs.Error) {
	rows, err := a.db.Query(`
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
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
			internalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			continue
		}

		appVersions = append(appVersions, appVersion)
	}

	return appVersions, nil
}

func (a AppVersion) CreateAppVersion(ct context.Context, appVersion entity.AppVersion) *errs.Error {
	// The uniqueness of version number is guaranteed by the primary key constraint of app_version table.
	// if try to insert into a new app version with the same version number, one should fail since db doesn't allow duplicate primary key
	// client could see according error and should retry
	_, err := a.db.Exec(`
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func (a AppVersion) UpdateAppVersion(ct context.Context, appVersion entity.AppVersion) *errs.Error {
	_, err := a.db.Exec(`
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func (a AppVersion) FindMaxVersionNumber(ct context.Context, appID uint64) (int32, *errs.Error) {
	var maxVersion int32
	err := a.db.QueryRow(`
	SELECT max(version_number)
	FROM app_version
	WHERE app_id = $1
`,
		appID,
	).Scan(&maxVersion)

	if errors.Is(err, sql.ErrNoRows) {
		// no version exists, start from 1
		return 0, nil
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return 0, internalErr
	}

	return maxVersion, nil
}

func (a AppVersion) DeleteAppVersion(ct context.Context, appID uint64, versionNumber int32) *errs.Error {
	_, err := a.db.Exec(`
		DELETE FROM app_version
		WHERE app_id = $1
		AND version_number = $2;
		`,
		appID,
		versionNumber)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func NewAppVersion(dataCollector telemetry.DataCollector, sqlDB *sql.DB) AppVersion {
	return AppVersion{dataCollector: dataCollector, db: sqlDB}
}
