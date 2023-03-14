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

type UserFileUploadSession struct {
	db *dbtest.InMemoryDB
}

var _ daov2.UserFileUploadSession = (*UserFileUploadSession)(nil)

func (u UserFileUploadSession) FindUserFileUploadSessionByUserIDWithTx(ct context.Context, tx *transaction.Transaction, userID uint64, userFileUploadSessionType entity.UserFileUploadSessionType, fileUploadSessionID uint64) (entity.UserFileUploadSession, *errs.Error) {
	var uploadSession entity.UserFileUploadSession
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := u.db.GetTable(UserFileUploadSessionTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currUploadSession := rawRow.(entity.UserFileUploadSession)
				if currUploadSession.UserID == userID &&
					currUploadSession.FileUploadSessionID == fileUploadSessionID &&
					currUploadSession.Type == userFileUploadSessionType {
					uploadSession = currUploadSession
					return nil
				}
			}

			return &errs.Error{
				Code:    errs.NotFound,
				Message: fmt.Sprintf("row not found: userID=%v, userFileUploadSessionType=%v, fileUploadSessionID=%v", userID, userFileUploadSessionType, fileUploadSessionID),
			}
		},
	})
	return uploadSession, err
}

func (u UserFileUploadSession) CreateUserFileUploadSession(ct context.Context, tx *transaction.Transaction, userFileUploadSession entity.UserFileUploadSession) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := u.db.GetTable(UserFileUploadSessionTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currUploadSession := row.(entity.UserFileUploadSession)
				if currUploadSession.UserID == userFileUploadSession.UserID &&
					currUploadSession.FileUploadSessionID == userFileUploadSession.FileUploadSessionID &&
					currUploadSession.Type == userFileUploadSession.Type {
					return errs.NewError(errs.Unknown, fmt.Sprintf("row already exist: userID=%v, userFileUploadSessionType=%v, fileUploadSessionID=%v",
						userFileUploadSession.UserID,
						userFileUploadSession.Type,
						userFileUploadSession.FileUploadSessionID))
				}
			}

			table.Rows = append(table.Rows, userFileUploadSession)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := u.db.GetTable(UserFileUploadSessionTableName)
			if err != nil {
				return err
			}

			table.Rows = table.Rows[:len(table.Rows)-1]
			return nil
		},
	})
}

func (u UserFileUploadSession) UpdateUserFileUploadSession(ct context.Context, tx *transaction.Transaction, userFileUploadSession entity.UserFileUploadSession) *errs.Error {
	var oldFileUploadSession entity.UserFileUploadSession
	oldFileUploadSessionFound := false
	table, err := u.db.GetTable(UserFileUploadSessionTableName)
	if err != nil {
		return err
	}

	for _, row := range table.Rows {
		currFileUploadSession := row.(entity.UserFileUploadSession)
		if currFileUploadSession.UserID == userFileUploadSession.UserID &&
			currFileUploadSession.FileUploadSessionID == userFileUploadSession.FileUploadSessionID &&
			currFileUploadSession.Type == userFileUploadSession.Type {
			oldFileUploadSession = currFileUploadSession
			oldFileUploadSessionFound = true
		}
	}

	if !oldFileUploadSessionFound {
		return errs.NewError(errs.Unknown, fmt.Sprintf("row not exist: userID=%v, userFileUploadSessionType=%v, fileUploadSessionID=%v",
			userFileUploadSession.UserID,
			userFileUploadSession.Type,
			userFileUploadSession.FileUploadSessionID))
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err = u.db.GetTable(UserFileUploadSessionTableName)
			if err != nil {
				return err
			}

			for i, row := range table.Rows {
				currFileUploadSession := row.(entity.UserFileUploadSession)
				if currFileUploadSession.UserID == userFileUploadSession.UserID &&
					currFileUploadSession.FileUploadSessionID == userFileUploadSession.FileUploadSessionID &&
					currFileUploadSession.Type == userFileUploadSession.Type {
					table.Rows[i] = userFileUploadSession
					return nil
				}
			}

			return errs.NewError(errs.Unknown, fmt.Sprintf("row not exist: userID=%v, userFileUploadSessionType=%v, fileUploadSessionID=%v",
				userFileUploadSession.UserID,
				userFileUploadSession.Type,
				userFileUploadSession.FileUploadSessionID))
		},
		Undo: func() *errs.Error {
			table, err = u.db.GetTable(UserFileUploadSessionTableName)
			if err != nil {
				return err
			}

			for i, row := range table.Rows {
				currFileUploadSession := row.(entity.UserFileUploadSession)
				if currFileUploadSession.UserID == userFileUploadSession.UserID &&
					currFileUploadSession.FileUploadSessionID == userFileUploadSession.FileUploadSessionID &&
					currFileUploadSession.Type == userFileUploadSession.Type {
					table.Rows[i] = oldFileUploadSession
				}
			}

			return nil
		},
	})
}

func NewUserFileUploadSession(db *dbtest.InMemoryDB) UserFileUploadSession {
	return UserFileUploadSession{db: db}
}
