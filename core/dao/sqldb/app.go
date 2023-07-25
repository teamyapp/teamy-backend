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

var _ dao.App = (*App)(nil)

type App struct {
	logger             telemetry.Logger
	transactionFactory transaction.Factory
}

func (a App) FindAllApps(ct context.Context) ([]entity.App, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return a.FindAllAppsWithTx(ct, tx)
}

func (a App) FindAppByID(ct context.Context, appID uint64) (entity.App, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.App{}, err
	}

	defer tx.Rollback()
	return a.FindAppByIDWithTx(ct, tx, appID)
}

func (a App) FindAppByIDWithTx(ct context.Context, tx *transaction.Transaction, appID uint64) (entity.App, *errs.Error) {
	app := entity.App{}
	err := tx.SQLTx().QueryRow(`
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

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.App{}, errs.NewError(errs.NotFound, fmt.Sprintf("app not found: appID=%v", appID))
		}

		return entity.App{}, errs.NewError(errs.Unknown, err.Error())
	}

	return app, nil
}

func (a App) FindAllAppsWithTx(ct context.Context, tx *transaction.Transaction) ([]entity.App, *errs.Error) {
	rows, err := tx.SQLTx().Query(`
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
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

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
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		apps = append(apps, app)
	}

	return apps, nil
}

func (a App) CreateApp(ct context.Context, tx *transaction.Transaction, app entity.App) *errs.Error {
	_, err := tx.SQLTx().Exec(`
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
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a App) UpdateApp(ct context.Context, tx *transaction.Transaction, app entity.App) *errs.Error {
	_, err := tx.SQLTx().Exec(`
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
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a App) DeleteApp(ct context.Context, tx *transaction.Transaction, appID uint64) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM app
		WHERE id = $1;
		`,
		appID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewApp(logger telemetry.Logger, transactionFactory transaction.Factory) App {
	return App{logger: logger, transactionFactory: transactionFactory}
}
