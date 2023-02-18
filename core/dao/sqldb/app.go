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

var _ dao.App = (*App)(nil)

type App struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

func (a App) FindAppByID(ct context.Context, appID uint64) (entity.App, *errs.Error) {
	app := entity.App{}
	err := a.db.QueryRow(`
	SELECT
	    id,
	    name,
	    description,
	    api_secret,
	    active_version_number,
	    installation_count,
	    creator_user_id,
	    created_at,
	    updated_at
	FROM app
	WHERE id = $1;
`,
		appID).
		Scan(
			&app.Name,
                         &app.Description,
			&app.ID,
			&app.APISecret,
			&app.ActiveVersionNumber,
			&app.InstallationCount,
			&app.CreatorUserID,
			&app.CreatedAt,
			&app.UpdatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("app not found: appID=%v", appID),
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.App{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.App{}, internalErr
	}

	return app, nil
}

func (a App) FindAllApps(ct context.Context) ([]entity.App, *errs.Error) {
	rows, err := a.db.Query(`
	SELECT
	    id,
	    apisecret,
	    active_version_number,
	    installation_count,
	    creator_user_id,
	    created_at,
	    updated_at,
	    description,
	    app_name
	FROM app
`)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	apps := make([]entity.App, 0)
	for rows.Next() {
		app := entity.App{}
		err = rows.Scan(
			&app.ID,
			&app.Name,
			&app.Description,
			&app.APISecret,
			&app.ActiveVersionNumber,
			&app.InstallationCount,
			&app.CreatorUserID,
			&app.CreatedAt,
			&app.UpdatedAt,
		)
		if err != nil {
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			a.dataCollector.Logger.ErrorWithContext(ct, newInternalErr)
			continue
		}

		apps = append(apps, app)
	}

	return apps, nil
}

func (a App) CreateApp(ct context.Context, app entity.App) *errs.Error {
	_, err := a.db.Exec(`
	INSERT INTO app
	(
	 	id,
	 	apisecret,
	 	active_version_number,
	 	installation_count,
	 	creator_user_id,
	 	created_at,
	 	updated_at,
	 	description,
	 	app_name
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`,
		app.ID,
		app.APISecret,
		app.ActiveVersionNumber,
		app.InstallationCount,
		app.CreatorUserID,
		app.CreatedAt,
		app.UpdatedAt,
		app.Description,
		app.AppName,
	)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (a App) UpdateApp(ct context.Context, app entity.App) *errs.Error {
	_, err := a.db.Exec(`
		UPDATE app
		SET
		    apisecret = $1,
		    active_version_number = $2,
		    installation_count = $3,
		    creator_user_id = $4,
		    created_at = $5,
		    updated_at = $6,
		    description = $7,
		    app_name = $8
		WHERE id = $9;`,
		app.ID,
		app.Name,
		app.Description,
		app.APISecret,
		app.ActiveVersionNumber,
		app.InstallationCount,
		app.CreatorUserID,
		app.CreatedAt,
		app.UpdatedAt,
	)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (a App) DeleteApp(ct context.Context, appID uint64) *errs.Error {
	_, err := a.db.Exec(`
		DELETE FROM app
		WHERE id = $1;
		`,
		appID)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func NewApp(dataCollector telemetry.DataCollector, sqlDB *sql.DB) App {
	return App{dataCollector: dataCollector, db: sqlDB}
}
