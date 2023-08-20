package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppPackageUploadSession struct {
	db *dbtest.InMemoryDB
}

var _ dao.AppPackageUploadSession = (*AppPackageUploadSession)(nil)

func (a *AppPackageUploadSession) FindAppPackageUploadSessionWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	appID uint64,
	userID uint64,
	versionNumber int32,
	fileUploadSessionID uint64,
) (entity.AppPackageUploadSession, *errs.Error) {
	var uploadSession entity.AppPackageUploadSession
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppPackageUploadSessionTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currUploadSession := rawRow.(entity.AppPackageUploadSession)
				if currUploadSession.AppID == appID &&
					currUploadSession.UserID == userID &&
					currUploadSession.VersionNumber == versionNumber &&
					currUploadSession.FileUploadSessionID == fileUploadSessionID {
					uploadSession = currUploadSession
					return nil
				}
			}

			return &errs.Error{
				Code:    errs.NotFound,
				Message: fmt.Sprint("row not found: appID=%v, userID=%v, versionNumber=%v, fileUploadSessionID=%v", appID, userID, versionNumber, fileUploadSessionID),
			}
		},
	})
	return uploadSession, err
}

func (a *AppPackageUploadSession) CreateAppPackageUploadSession(ct context.Context, tx *transaction.Transaction, session entity.AppPackageUploadSession) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppPackageUploadSessionTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currUploadSession := rawRow.(entity.AppPackageUploadSession)
				if currUploadSession.AppID == appPackageUploadSession.AppID &&
					currUploadSession.UserID == appPackageUploadSession.UserID &&
					currUploadSession.VersionNumber == appPackageUploadSession.VersionNumber {
					return &errs.Error{
						Code:    errs.AlreadyExists,
						Message: fmt.Sprint("row already exists: appID=%v, userID=%v, versionNumber=%v", appPackageUploadSession.AppID, appPackageUploadSession.UserID, appPackageUploadSession.VersionNumber),
					}
				}
			}

			table.Rows = append(table.Rows, appPackageUploadSession)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := a.db.GetTable(AppPackageUploadSessionTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, rawRow := range table.Rows {
				currUploadSession := rawRow.(entity.AppPackageUploadSession)
				if currUploadSession.AppID == appPackageUploadSession.AppID &&
					currUploadSession.UserID == appPackageUploadSession.UserID &&
					currUploadSession.VersionNumber == appPackageUploadSession.VersionNumber {
					continue
				}

				rows = append(rows, rawRow)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (a *AppPackageUploadSession) UpdateAppPackageFileUploadSession(
    ct context.Context, 
    tx *transaction.Transaction, 
    session entity.AppPackageUploadSession,
) *errs.Error {
	oldAppPackageUploadSession, err := a.FindAppPackageUploadSessionWithTx(ct, tx, appPackageUploadSession.AppID, appPackageUploadSession.UserID, appPackageUploadSession.VersionNumber, appPackageUploadSession.FileUploadSessionID)
	if err != nil {
		return err
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppPackageUploadSessionTableName)
			if err != nil {
				return err
			}

			for index, rawRow := range table.Rows {
				currUploadSession := rawRow.(entity.AppPackageUploadSession)
				if currUploadSession.AppID == appPackageUploadSession.AppID &&
					currUploadSession.UserID == appPackageUploadSession.UserID &&
					currUploadSession.VersionNumber == appPackageUploadSession.VersionNumber {
					table.Rows[i] = appPackageUploadSession
					return nil
				}
			}

			return &errs.Error{
				Code: errs.NotFound,
				Message: fmt.Sprint("row not found: appID=%v, userID=%v, versionNumber=%v",
					appPackageUploadSession.AppID,
					appPackageUploadSession.UserID,
					appPackageUploadSession.VersionNumber),
			}
		},
		Undo: func() *errs.Error {
			table, err := a.db.GetTable(AppPackageUploadSessionTableName)
			if err != nil {
				return err
			}

			for index, rawRow := range table.Rows {
				currUploadSession := rawRow.(entity.AppPackageUploadSession)
				if currUploadSession.AppID == appPackageUploadSession.AppID &&
					currUploadSession.UserID == appPackageUploadSession.UserID &&
					currUploadSession.VersionNumber == appPackageUploadSession.VersionNumber {
					table.Rows[i] = oldAppPackageUploadSession
					return nil
				}
			}

			return &errs.Error{
				Code: errs.NotFound,
				Message: fmt.Sprint("row not found: appID=%v, userID=%v, versionNumber=%v",
					appPackageUploadSession.AppID,
					appPackageUploadSession.UserID,
					appPackageUploadSession.VersionNumber),
			}
		},
	})
}

func NewAppPackageUploadSession(db *dbtest.InMemoryDB) *AppPackageUploadSession {
	return &AppPackageUploadSession{
		db: db,
	}
}
