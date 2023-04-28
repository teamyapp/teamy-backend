package daotestv2

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type App struct {
	db                 *dbtest.InMemoryDB
	transactionFactory transaction.Factory
}

var _ daov2.App = (*App)(nil)

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

func (a App) FindAppByIDWithTx(ct context.Context, tx *transaction.Transaction, appID uint64) (entity.App, *errs.Error) {
	var app entity.App
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currApp := rawRow.(entity.App)
				if currApp.ID == appID {
					app = currApp
					return nil
				}
			}

			return errs.NewError(errs.NotFound, fmt.Sprintf("row not found: appID=%v", appID))
		},
	})
	return app, err
}

func (a App) FindAllAppsWithTx(ct context.Context, tx *transaction.Transaction) ([]entity.App, *errs.Error) {
	var apps []entity.App
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currApp := rawRow.(entity.App)
				apps = append(apps, currApp)
			}

			return nil
		},
	})
	return apps, err
}

func (a App) CreateApp(ct context.Context, tx *transaction.Transaction, app entity.App) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currApp := row.(entity.App)
				if currApp.ID == app.ID {
					return errs.NewError(errs.Unknown, fmt.Sprintf("row already exist: appID=%v", app.ID))
				}
			}

			table.Rows = append(table.Rows, app)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := a.db.GetTable(AppTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currApp := row.(entity.App)
				if currApp.ID == app.ID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (a App) UpdateApp(ct context.Context, tx *transaction.Transaction, app entity.App) *errs.Error {
	oldApp, internalErr := a.FindAppByIDWithTx(ct, tx, app.ID)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppTableName)
			if err != nil {
				return err
			}

			for i, row := range table.Rows {
				currApp := row.(entity.App)
				if currApp.ID == app.ID {
					table.Rows[i] = app
					return nil
				}
			}

			return errs.NewError(errs.Unknown, fmt.Sprintf("row not exist: appID=%v", app.ID))
		},
		Undo: func() *errs.Error {
			table, err := a.db.GetTable(AppTableName)
			if err != nil {
				return err
			}

			for index, row := range table.Rows {
				currApp := row.(entity.App)
				if currApp.ID == app.ID {
					table.Rows[index] = oldApp
				}
			}

			return nil
		},
	})
}

func (a App) DeleteApp(ct context.Context, tx *transaction.Transaction, appID uint64) *errs.Error {
	oldApp, internalErr := a.FindAppByIDWithTx(ct, tx, appID)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currApp := row.(entity.App)
				if currApp.ID != appID {
					rows = append(rows, currApp)
				}
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := a.db.GetTable(AppTableName)
			if err != nil {
				return err
			}

			table.Rows = append(table.Rows, oldApp)
			return nil
		},
	})
}

func NewApp(db *dbtest.InMemoryDB, transactionFactory transaction.Factory) App {
	return App{db: db, transactionFactory: transactionFactory}
}
