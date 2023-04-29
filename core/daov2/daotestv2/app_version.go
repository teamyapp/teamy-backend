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

type AppVersion struct {
	db                 *dbtest.InMemoryDB
	transactionFactory transaction.Factory
}

var _ daov2.AppVersion = (*AppVersion)(nil)

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

func (a AppVersion) FindAppVersionByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int32) (entity.AppVersion, *errs.Error) {
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

func (a AppVersion) FindAppVersionByAppIDAndVersionNumberWithTx(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int32) (entity.AppVersion, *errs.Error) {
	var appVersion entity.AppVersion
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppVersionTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currAppVersion := rawRow.(entity.AppVersion)
				if currAppVersion.AppID == appID &&
					currAppVersion.VersionNumber == versionNumber {
					appVersion = currAppVersion
					return nil
				}
			}

			return errs.NewError(errs.NotFound, fmt.Sprintf("row not found: appID=%v, versionNumber=%v", appID, versionNumber))
		},
	})
	return appVersion, err
}

func (a AppVersion) FindAppVersionsByAppIDWithTx(ct context.Context, tx *transaction.Transaction, appID uint64) ([]entity.AppVersion, *errs.Error) {
	var appVersions []entity.AppVersion
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppVersionTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currAppVersion := rawRow.(entity.AppVersion)
				if currAppVersion.AppID == appID {
					appVersions = append(appVersions, currAppVersion)
				}
			}

			return nil
		},
	})
	return appVersions, err
}

func (a AppVersion) FindMaxVersionNumberWithTx(ct context.Context, tx *transaction.Transaction, appID uint64) (int32, *errs.Error) {
	maxVersionNumber := int32(0)
	rowFound := false
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppVersionTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currAppVersion := rawRow.(entity.AppVersion)
				if currAppVersion.AppID == appID {
					rowFound = true
					if currAppVersion.VersionNumber > maxVersionNumber {
						maxVersionNumber = currAppVersion.VersionNumber
					}
				}
			}

			if !rowFound {
				return errs.NewError(errs.NotFound, fmt.Sprintf("row not found: appID=%v", appID))
			}

			return nil
		},
	})
	return maxVersionNumber, err
}

func (a AppVersion) CreateAppVersion(ct context.Context, tx *transaction.Transaction, appVersion entity.AppVersion) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppVersionTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currAppVersion := row.(entity.AppVersion)
				if currAppVersion.AppID == appVersion.AppID &&
					currAppVersion.VersionNumber == appVersion.VersionNumber {
					return errs.NewError(errs.Unknown, fmt.Sprintf("row already exist: appID=%v, versionNumber=%v", appVersion.AppID, appVersion.VersionNumber))
				}
			}

			table.Rows = append(table.Rows, appVersion)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := a.db.GetTable(AppVersionTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currAppVersion := row.(entity.AppVersion)
				if currAppVersion.AppID == appVersion.AppID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (a AppVersion) UpdateAppVersion(ct context.Context, tx *transaction.Transaction, appVersion entity.AppVersion) *errs.Error {
	oldApp, internalErr := a.FindAppVersionByAppIDAndVersionNumberWithTx(ct, tx, appVersion.AppID, appVersion.VersionNumber)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppVersionTableName)
			if err != nil {
				return err
			}

			for i, row := range table.Rows {
				currAppVersion := row.(entity.AppVersion)
				if currAppVersion.AppID == appVersion.AppID &&
					currAppVersion.VersionNumber == appVersion.VersionNumber {
					table.Rows[i] = appVersion
					return nil
				}
			}

			return errs.NewError(errs.Unknown, fmt.Sprintf("row not exist: appID=%v, versionNumber=%v", appVersion.AppID, appVersion.VersionNumber))
		},
		Undo: func() *errs.Error {
			table, err := a.db.GetTable(AppVersionTableName)
			if err != nil {
				return err
			}

			for index, row := range table.Rows {
				currAppVersion := row.(entity.AppVersion)
				if currAppVersion.AppID == appVersion.AppID {
					table.Rows[index] = oldApp
				}
			}

			return nil
		},
	})
}

func (a AppVersion) DeleteAppVersion(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int32) *errs.Error {
	oldAppVersion, internalErr := a.FindAppVersionByAppIDAndVersionNumberWithTx(ct, tx, appID, versionNumber)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppVersionTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currAppVersion := row.(entity.AppVersion)
				if currAppVersion.AppID != appID ||
					currAppVersion.VersionNumber != versionNumber {
					rows = append(rows, currAppVersion)
				}
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := a.db.GetTable(AppVersionTableName)
			if err != nil {
				return err
			}

			table.Rows = append(table.Rows, oldAppVersion)
			return nil
		},
	})
}

func NewAppVersion(db *dbtest.InMemoryDB, transactionFactory transaction.Factory) AppVersion {
	return AppVersion{db: db, transactionFactory: transactionFactory}
}
