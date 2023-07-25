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

type TeamFileUploadSession struct {
	db *dbtest.InMemoryDB
}

var _ dao.TeamFileUploadSession = (*TeamFileUploadSession)(nil)

func (t TeamFileUploadSession) FindTeamFileUploadSessionByTeamIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	teamID uint64,
	teamFileUploadSessionType entity.TeamFileUploadSessionType,
	fileUploadSessionID uint64,
) (entity.TeamFileUploadSession, *errs.Error) {
	var uploadSession entity.TeamFileUploadSession
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamFileUploadSessionTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currUploadSession := rawRow.(entity.TeamFileUploadSession)
				if currUploadSession.TeamID == teamID &&
					currUploadSession.FileUploadSessionID == fileUploadSessionID &&
					currUploadSession.Type == teamFileUploadSessionType {
					uploadSession = currUploadSession
					return nil
				}
			}

			return &errs.Error{
				Code:    errs.NotFound,
				Message: fmt.Sprintf("row not found: teamID=%v, teamFileUploadSessionType=%v, fileUploadSessionID=%v", teamID, teamFileUploadSessionType, fileUploadSessionID),
			}
		},
	})
	return uploadSession, err
}

func (t TeamFileUploadSession) CreateTeamFileUploadSession(ct context.Context, tx *transaction.Transaction, teamFileUploadSession entity.TeamFileUploadSession) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamFileUploadSessionTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currUploadSession := row.(entity.TeamFileUploadSession)
				if currUploadSession.TeamID == teamFileUploadSession.TeamID &&
					currUploadSession.FileUploadSessionID == teamFileUploadSession.FileUploadSessionID &&
					currUploadSession.Type == teamFileUploadSession.Type {
					return errs.NewError(errs.Unknown, fmt.Sprintf("row already exist: teamID=%v, teamFileUploadSessionType=%v, fileUploadSessionID=%v",
						teamFileUploadSession.TeamID,
						teamFileUploadSession.Type,
						teamFileUploadSession.FileUploadSessionID))
				}
			}

			table.Rows = append(table.Rows, teamFileUploadSession)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TeamFileUploadSessionTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currUserFileUploadSession := row.(entity.TeamFileUploadSession)
				if currUserFileUploadSession.FileUploadSessionID == teamFileUploadSession.FileUploadSessionID &&
					currUserFileUploadSession.TeamID == teamFileUploadSession.TeamID &&
					currUserFileUploadSession.Type == teamFileUploadSession.Type {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (t TeamFileUploadSession) UpdateTeamFileUploadSession(ct context.Context, tx *transaction.Transaction, teamFileUploadSession entity.TeamFileUploadSession) *errs.Error {
	oldFileUploadSession, internalErr := t.FindTeamFileUploadSessionByTeamIDWithTx(ct, tx, teamFileUploadSession.TeamID, teamFileUploadSession.Type, teamFileUploadSession.FileUploadSessionID)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamFileUploadSessionTableName)
			if err != nil {
				return err
			}

			for index, row := range table.Rows {
				currFileUploadSession := row.(entity.TeamFileUploadSession)
				if currFileUploadSession.TeamID == teamFileUploadSession.TeamID &&
					currFileUploadSession.FileUploadSessionID == teamFileUploadSession.FileUploadSessionID &&
					currFileUploadSession.Type == teamFileUploadSession.Type {
					table.Rows[index] = teamFileUploadSession
					return nil
				}
			}

			return errs.NewError(errs.Unknown, fmt.Sprintf("row not exist: teamID=%v, teamFileUploadSessionType=%v, fileUploadSessionID=%v",
				teamFileUploadSession.TeamID,
				teamFileUploadSession.Type,
				teamFileUploadSession.FileUploadSessionID))
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TeamFileUploadSessionTableName)
			if err != nil {
				return err
			}

			for index, row := range table.Rows {
				currFileUploadSession := row.(entity.TeamFileUploadSession)
				if currFileUploadSession.TeamID == teamFileUploadSession.TeamID &&
					currFileUploadSession.FileUploadSessionID == teamFileUploadSession.FileUploadSessionID &&
					currFileUploadSession.Type == teamFileUploadSession.Type {
					table.Rows[index] = oldFileUploadSession
					return nil
				}
			}

			return nil
		},
	})
}

func NewTeamFileUploadSession(db *dbtest.InMemoryDB) TeamFileUploadSession {
	return TeamFileUploadSession{db: db}
}
