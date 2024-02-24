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

func (a *AppPackageUploadSession) FindAppPackageUploadSessionByAppIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	appID uint64,
) ([]entity.AppPackageUploadSession, *errs.Error) {
	var uploadSessions []entity.AppPackageUploadSession
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppPackageUploadSessionTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currUploadSession := rawRow.(entity.AppPackageUploadSession)
				if currUploadSession.AppID == appID {
					uploadSessions = append(uploadSessions, currUploadSession)
				}
			}

			return nil
		},
	})
	return uploadSessions, err
}

func (a *AppPackageUploadSession) FindAppPackageUploadSessionWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	appID uint64,
	versionNumber int,
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
					currUploadSession.VersionNumber == versionNumber &&
					currUploadSession.FileUploadSessionID == fileUploadSessionID {
					uploadSession = currUploadSession
					return nil
				}
			}

			return &errs.Error{
				Code:    errs.NotFound,
				Message: fmt.Sprintf("row not found: appID=%v, versionNumber=%v, fileUploadSessionID=%v", appID, versionNumber, fileUploadSessionID),
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
				if currUploadSession.AppID == session.AppID &&
					currUploadSession.VersionNumber == session.VersionNumber {
					return &errs.Error{
						Code:    errs.AlreadyExists,
						Message: fmt.Sprintf("row already exists: appID=%v, versionNumber=%v", session.AppID, session.VersionNumber),
					}
				}
			}

			table.Rows = append(table.Rows, session)
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
				if currUploadSession.AppID == session.AppID &&
					currUploadSession.VersionNumber == session.VersionNumber {
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
	oldAppPackageUploadSession, err := a.FindAppPackageUploadSessionWithTx(ct, tx, session.AppID, session.VersionNumber, session.FileUploadSessionID)
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
				if currUploadSession.AppID == session.AppID &&
					currUploadSession.VersionNumber == session.VersionNumber {
					table.Rows[index] = session
					return nil
				}
			}

			return &errs.Error{
				Code: errs.NotFound,
				Message: fmt.Sprintf("row not found: appID=%v, versionNumber=%v",
					session.AppID,
					session.VersionNumber),
			}
		},
		Undo: func() *errs.Error {
			table, err := a.db.GetTable(AppPackageUploadSessionTableName)
			if err != nil {
				return err
			}

			for index, rawRow := range table.Rows {
				currUploadSession := rawRow.(entity.AppPackageUploadSession)
				if currUploadSession.AppID == session.AppID &&
					currUploadSession.VersionNumber == session.VersionNumber {
					table.Rows[index] = oldAppPackageUploadSession
					return nil
				}
			}

			return &errs.Error{
				Code: errs.NotFound,
				Message: fmt.Sprintf("row not found: appID=%v, versionNumber=%v",
					session.AppID,
					session.VersionNumber),
			}
		},
	})
}

func (a *AppPackageUploadSession) DeleteAppPackageUploadSessionsByAppID(ct context.Context, tx *transaction.Transaction, appID uint64) *errs.Error {
	oldAppPackageUploadSessions, err := a.FindAppPackageUploadSessionByAppIDWithTx(ct, tx, appID)
	if err != nil {
		return err
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := a.db.GetTable(AppPackageUploadSessionTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, rawRow := range table.Rows {
				currUploadSession := rawRow.(entity.AppPackageUploadSession)
				if currUploadSession.AppID == appID {
					continue
				}

				rows = append(rows, rawRow)
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := a.db.GetTable(AppPackageUploadSessionTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currUploadSession := rawRow.(entity.AppPackageUploadSession)
				if currUploadSession.AppID == appID {
					return &errs.Error{
						Code: errs.Unknown,
						Message: fmt.Sprintf("row already exists: appID=%v",
							appID),
					}
				}
			}

			for _, uploadSession := range oldAppPackageUploadSessions {
				table.Rows = append(table.Rows, uploadSession)
			}

			return nil
		},
	})
}

func NewAppPackageUploadSession(db *dbtest.InMemoryDB) *AppPackageUploadSession {
	return &AppPackageUploadSession{
		db: db,
	}
}
