package daotestv2

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppVersionVisibleTeam struct {
	db                 *dbtest.InMemoryDB
	transactionFactory transaction.Factory
}

var _ daov2.AppVersionVisibleTeam = (*AppVersionVisibleTeam)(nil)

func (a AppVersionVisibleTeam) FindAppVersionVisibleTeamWithTx(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int32, teamID uint64) (entity.AppVersionVisibleTeam, *errs.Error) {
	var appVersionVisibleTeam entity.AppVersionVisibleTeam
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppVersionVisibleTeamTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currAppVersionVisibleTeam := rawRow.(entity.AppVersionVisibleTeam)
				if currAppVersionVisibleTeam.AppID == appID &&
					currAppVersionVisibleTeam.VersionNumber == versionNumber &&
					currAppVersionVisibleTeam.TeamID == teamID {
					appVersionVisibleTeam = currAppVersionVisibleTeam
					return nil
				}
			}

			return errs.NewError(errs.NotFound, fmt.Sprintf("row not found: appID=%v, versionNumber=%v, teamID=%v",
				appID, versionNumber, teamID))
		},
	})
	return appVersionVisibleTeam, err
}

func (a AppVersionVisibleTeam) FindAppVersionVisibleTeamsByAppIDAndVersionNumberWithTx(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int32) ([]entity.AppVersionVisibleTeam, *errs.Error) {
	var appVersionVisibleTeams []entity.AppVersionVisibleTeam
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppVersionVisibleTeamTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currAppVersionVisibleTeam := rawRow.(entity.AppVersionVisibleTeam)
				if currAppVersionVisibleTeam.AppID == appID && currAppVersionVisibleTeam.VersionNumber == versionNumber {
					appVersionVisibleTeams = append(appVersionVisibleTeams, currAppVersionVisibleTeam)
				}
			}

			return nil
		},
	})
	return appVersionVisibleTeams, err
}

func (a AppVersionVisibleTeam) FindAppVersionVisibleTeamsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.AppVersionVisibleTeam, *errs.Error) {
	var appVersionVisibleTeams []entity.AppVersionVisibleTeam
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppVersionVisibleTeamTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currAppVersionVisibleTeam := rawRow.(entity.AppVersionVisibleTeam)
				if currAppVersionVisibleTeam.TeamID == teamID {
					appVersionVisibleTeams = append(appVersionVisibleTeams, currAppVersionVisibleTeam)
				}
			}

			return nil
		},
	})
	return appVersionVisibleTeams, err
}

func (a AppVersionVisibleTeam) CreateAppVersionVisibleTeam(ct context.Context, tx *transaction.Transaction, appVersionVisibleTeam entity.AppVersionVisibleTeam) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppVersionVisibleTeamTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currAppVersionVisibleTeam := row.(entity.AppVersionVisibleTeam)
				if currAppVersionVisibleTeam.AppID == appVersionVisibleTeam.AppID &&
					currAppVersionVisibleTeam.VersionNumber == appVersionVisibleTeam.VersionNumber &&
					currAppVersionVisibleTeam.TeamID == appVersionVisibleTeam.TeamID {
					return errs.NewError(errs.Unknown, fmt.Sprintf("row already exist: appID=%v, versionNumber=%v, teamID=%v",
						appVersionVisibleTeam.AppID, appVersionVisibleTeam.VersionNumber, appVersionVisibleTeam.TeamID))
				}
			}

			table.Rows = append(table.Rows, appVersionVisibleTeam)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := a.db.GetTable(AppVersionVisibleTeamTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currAppVersionVisibleTeam := row.(entity.AppVersionVisibleTeam)
				if currAppVersionVisibleTeam.AppID == appVersionVisibleTeam.AppID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (a AppVersionVisibleTeam) DeleteAppVersionVisibleTeam(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int32, teamID uint64) *errs.Error {
	oldAppVersionVisibleTeam, internalErr := a.FindAppVersionVisibleTeamWithTx(ct, tx, appID, versionNumber, teamID)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppVersionVisibleTeamTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currAppVersionVisibleTeam := row.(entity.AppVersionVisibleTeam)
				if currAppVersionVisibleTeam.AppID != appID &&
					currAppVersionVisibleTeam.TeamID == teamID &&
					currAppVersionVisibleTeam.VersionNumber == versionNumber {
					rows = append(rows, currAppVersionVisibleTeam)
				}
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := a.db.GetTable(AppVersionVisibleTeamTableName)
			if err != nil {
				return err
			}

			table.Rows = append(table.Rows, oldAppVersionVisibleTeam)
			return nil
		},
	})
}

func (a AppVersionVisibleTeam) DeleteAppVersionVisibleTeamsByAppIDAndVersionNumber(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int32) *errs.Error {
	var oldAppVersionVisibleTeams []entity.AppVersionVisibleTeam
	table, err := a.db.GetTable(AppVersionVisibleTeamTableName)
	if err != nil {
		return nil
	}
	for _, row := range table.Rows {
		currAppVersionVisibleTeam := row.(entity.AppVersionVisibleTeam)
		if currAppVersionVisibleTeam.AppID == appID &&
			currAppVersionVisibleTeam.VersionNumber == versionNumber {
			oldAppVersionVisibleTeams = append(oldAppVersionVisibleTeams, currAppVersionVisibleTeam)
		}
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err = a.db.GetTable(AppVersionVisibleTeamTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currAppVersionVisibleTeam := row.(entity.AppVersionVisibleTeam)
				if currAppVersionVisibleTeam.AppID != appID ||
					currAppVersionVisibleTeam.VersionNumber != versionNumber {
					rows = append(rows, currAppVersionVisibleTeam)
				}
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err = a.db.GetTable(AppVersionVisibleTeamTableName)
			if err != nil {
				return err
			}

			table.Rows = append(table.Rows, oldAppVersionVisibleTeams)
			return nil
		},
	})
}

func NewAppVersionVisibleTeam(db *dbtest.InMemoryDB, transactionFactory transaction.Factory) AppVersionVisibleTeam {
	return AppVersionVisibleTeam{db: db, transactionFactory: transactionFactory}
}
