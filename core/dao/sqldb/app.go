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
	logger telemetry.Logger
	db     *sql.DB
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

	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("app not found: appID=%v", appID),
		}
		a.logger.ErrorWithContext(ct, internalErr)
		return entity.App{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.logger.ErrorWithContext(ct, internalErr)
		return entity.App{}, internalErr
	}

	return app, nil
}

func (a App) FindAllApps(ct context.Context) ([]entity.App, *errs.Error) {
	rows, err := a.db.Query(`
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
`)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.logger.ErrorWithContext(ct, internalErr)
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

			a.logger.ErrorWithContext(ct, newInternalErr)
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
	 	name,
	 	description,
	 	api_secret,
	 	active_version_number,
	 	installation_count,
	 	creator_user_id,
	 	created_at,
	 	updated_at
	)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`,
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
		a.logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (a App) UpdateApp(ct context.Context, app entity.App) *errs.Error {
	_, err := a.db.Exec(`
		UPDATE app
		SET
		    name = $1,
		    description = $2,
		    api_secret = $3,
		    active_version_number = $4,
		    installation_count = $5,
		    creator_user_id = $6,
		    created_at = $7,
		    updated_at = $8
		WHERE id = $9;`,
		app.Name,
		app.Description,
		app.APISecret,
		app.ActiveVersionNumber,
		app.InstallationCount,
		app.CreatorUserID,
		app.CreatedAt,
		app.UpdatedAt,
		app.ID,
	)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.logger.ErrorWithContext(ct, internalErr)
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
		a.logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func NewApp(logger telemetry.Logger, sqlDB *sql.DB) App {
	return App{logger: logger, db: sqlDB}
}
