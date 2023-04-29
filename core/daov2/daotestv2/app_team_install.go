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

type AppTeamInstallation struct {
	db                 *dbtest.InMemoryDB
	transactionFactory transaction.Factory
}

var _ daov2.AppTeamInstallation = (*AppTeamInstallation)(nil)

func (a AppTeamInstallation) FindAppTeamInstallationsByAppID(ct context.Context, appID uint64) ([]entity.AppTeamInstallation, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return a.FindAppTeamInstallationsByAppIDWithTx(ct, tx, appID)
}

func (a AppTeamInstallation) FindAppTeamInstallationsByTeamID(ct context.Context, teamID uint64) ([]entity.AppTeamInstallation, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return a.FindAppTeamInstallationsByTeamIDWithTx(ct, tx, teamID)
}

func (a AppTeamInstallation) FindAppTeamInstallationsByAppIDWithTx(ct context.Context, tx *transaction.Transaction, appID uint64) ([]entity.AppTeamInstallation, *errs.Error) {
	var appTeamInstallations []entity.AppTeamInstallation
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppTeamInstallationTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currAppTeamInstallation := rawRow.(entity.AppTeamInstallation)
				if currAppTeamInstallation.AppID == appID {
					appTeamInstallations = append(appTeamInstallations, currAppTeamInstallation)
				}
			}

			return nil
		},
	})
	return appTeamInstallations, err
}

func (a AppTeamInstallation) FindAppTeamInstallationsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.AppTeamInstallation, *errs.Error) {
	var appTeamInstallations []entity.AppTeamInstallation
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppTeamInstallationTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currAppTeamInstallation := rawRow.(entity.AppTeamInstallation)
				if currAppTeamInstallation.InstalledTeamID == teamID {
					appTeamInstallations = append(appTeamInstallations, currAppTeamInstallation)
				}
			}

			return nil
		},
	})
	return appTeamInstallations, err
}

func (a AppTeamInstallation) FindAppTeamInstallationByAppIDAndTeamIDWithTx(ct context.Context, tx *transaction.Transaction, appID uint64, teamID uint64) (entity.AppTeamInstallation, *errs.Error) {
	var appTeamInstallation entity.AppTeamInstallation
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppTeamInstallationTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currAppTeamInstallation := rawRow.(entity.AppTeamInstallation)
				if currAppTeamInstallation.AppID == appID &&
					currAppTeamInstallation.InstalledTeamID == teamID {
					appTeamInstallation = currAppTeamInstallation
					return nil
				}
			}

			return errs.NewError(errs.NotFound, fmt.Sprintf("row not found: appID=%v, teamID=%v", appID, teamID))
		},
	})
	return appTeamInstallation, err
}

func (a AppTeamInstallation) CreateAppTeamInstallation(ct context.Context, tx *transaction.Transaction, appTeamInstallation entity.AppTeamInstallation) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppTeamInstallationTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currAppTeamInstallation := row.(entity.AppTeamInstallation)
				if currAppTeamInstallation.AppID == appTeamInstallation.AppID &&
					currAppTeamInstallation.InstalledTeamID == appTeamInstallation.InstalledTeamID {
					return errs.NewError(errs.Unknown,
						fmt.Sprintf("row already exist: appID=%v, teamID=%v", appTeamInstallation.AppID, appTeamInstallation.InstalledTeamID))
				}
			}

			table.Rows = append(table.Rows, appTeamInstallation)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := a.db.GetTable(AppTeamInstallationTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currAppTeamInstallation := row.(entity.AppTeamInstallation)
				if currAppTeamInstallation.AppID == appTeamInstallation.AppID &&
					currAppTeamInstallation.InstalledTeamID == appTeamInstallation.InstalledTeamID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (a AppTeamInstallation) UpdateAppTeamInstallation(ct context.Context, tx *transaction.Transaction, appTeamInstallation entity.AppTeamInstallation) *errs.Error {
	oldAppTeamInstallation, internalErr := a.FindAppTeamInstallationByAppIDAndTeamIDWithTx(ct, tx,
		appTeamInstallation.AppID, appTeamInstallation.InstalledTeamID)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppTeamInstallationTableName)
			if err != nil {
				return err
			}

			for i, row := range table.Rows {
				currAppTeamInstallation := row.(entity.AppTeamInstallation)
				if currAppTeamInstallation.AppID == appTeamInstallation.AppID &&
					currAppTeamInstallation.InstalledTeamID == appTeamInstallation.InstalledTeamID {
					table.Rows[i] = appTeamInstallation
					return nil
				}
			}

			return errs.NewError(errs.Unknown, fmt.Sprintf("row not exist: appID=%v, teamID=%v",
				appTeamInstallation.AppID, appTeamInstallation.InstalledTeamID))
		},
		Undo: func() *errs.Error {
			table, err := a.db.GetTable(AppTeamInstallationTableName)
			if err != nil {
				return err
			}

			for index, row := range table.Rows {
				currAppTeamInstallation := row.(entity.AppTeamInstallation)
				if currAppTeamInstallation.AppID == appTeamInstallation.AppID &&
					currAppTeamInstallation.InstalledTeamID == appTeamInstallation.InstalledTeamID {
					table.Rows[index] = oldAppTeamInstallation
				}
			}

			return nil
		},
	})
}

func (a AppTeamInstallation) DeleteAppTeamInstallation(ct context.Context, tx *transaction.Transaction, appID uint64, teamID uint64) *errs.Error {
	oldAppTeamInstallation, internalErr := a.FindAppTeamInstallationByAppIDAndTeamIDWithTx(ct, tx, appID, teamID)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppTeamInstallationTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currAppTeamInstallation := row.(entity.AppTeamInstallation)
				if currAppTeamInstallation.AppID != appID &&
					currAppTeamInstallation.InstalledTeamID == teamID {
					rows = append(rows, currAppTeamInstallation)
				}
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := a.db.GetTable(AppTeamInstallationTableName)
			if err != nil {
				return err
			}

			table.Rows = append(table.Rows, oldAppTeamInstallation)
			return nil
		},
	})
}

func (a AppTeamInstallation) DeleteAppTeamInstallationsByAppIDAndVersionNumber(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int32) *errs.Error {
	var oldAppTeamInstallations []entity.AppTeamInstallation
	table, err := a.db.GetTable(AppTeamInstallationTableName)
	if err != nil {
		return nil
	}
	for _, row := range table.Rows {
		currAppTeamInstallation := row.(entity.AppTeamInstallation)
		if currAppTeamInstallation.AppID == appID &&
			currAppTeamInstallation.EnabledVersionNumber == versionNumber {
			oldAppTeamInstallations = append(oldAppTeamInstallations, currAppTeamInstallation)
		}
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err = a.db.GetTable(AppTeamInstallationTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currAppTeamInstallation := row.(entity.AppTeamInstallation)
				if currAppTeamInstallation.AppID != appID ||
					currAppTeamInstallation.EnabledVersionNumber == versionNumber {
					rows = append(rows, currAppTeamInstallation)
				}
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err = a.db.GetTable(AppTeamInstallationTableName)
			if err != nil {
				return err
			}

			table.Rows = append(table.Rows, oldAppTeamInstallations)
			return nil
		},
	})
}

func NewAppTeamInstallation(db *dbtest.InMemoryDB, transactionFactory transaction.Factory) AppTeamInstallation {
	return AppTeamInstallation{db: db, transactionFactory: transactionFactory}
}
