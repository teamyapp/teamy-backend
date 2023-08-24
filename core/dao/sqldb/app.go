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

type App struct {
	logger             telemetry.Logger
	transactionFactory transaction.Factory
}

var _ dao.App = (*App)(nil)

func (a *App) FindAppByID(ct context.Context, appID uint64) (entity.App, *errs.Error) {
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

func (a *App) FindAppByIDWithTx(ct context.Context, tx *transaction.Transaction, appID uint64) (entity.App, *errs.Error) {
	app := entity.App{}
	err := tx.SQLTx().QueryRow(`
		SELECT
			id,
			total_installations,
			created_at,
			updated_at,
			managed_by_team_id
		FROM app
		WHERE id = $1;`,
		appID).
		Scan(
			&app.ID,
			&app.TotalInstallations,
			&app.CreatedAt,
			&app.UpdatedAt,
			&app.ManagedByTeamID,
		)
	if err != nil {
		return entity.App{}, errs.NewError(errs.Unknown, err.Error())
	}

	return app, nil
}

func (a *App) CreateApp(ct context.Context, tx *transaction.Transaction, app entity.App) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		INSERT INTO app (
			id,
			total_installations,
			created_at,
			updated_at,
			managed_by_team_id
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5
		);`,
		app.ID,
		app.TotalInstallations,
		app.CreatedAt,
		app.UpdatedAt,
		app.ManagedByTeamID,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a *App) DeleteApp(ct context.Context, tx *transaction.Transaction, appID uint64) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM app
		WHERE id = $1;`,
		appID,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewApp(logger telemetry.Logger, transactionFactory transaction.Factory) *App {
	return &App{
		logger:             logger,
		transactionFactory: transactionFactory,
	}
}
