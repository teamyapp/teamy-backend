package daotest

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamMember struct {
	db                 *dbtest.InMemoryDB
	transactionFactory transaction.Factory
}

var _ dao.TeamMember = (*TeamMember)(nil)

func (t TeamMember) FindTeamMembersByTeamID(ct context.Context, teamID uint64) ([]entity.TeamMember, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := t.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return t.FindTeamMembersByTeamIDWithTx(ct, tx, teamID)
}

func (t TeamMember) FindTeamIDsByUserID(ct context.Context, userID uint64) ([]uint64, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := t.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return t.FindTeamIDsByUserIDWithTx(ct, tx, userID)
}

func (t TeamMember) FindTeamIDsByUserIDWithTx(ct context.Context, tx *transaction.Transaction, userID uint64) ([]uint64, *errs.Error) {
	var teamIDs []uint64
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currTeamMember := rawRow.(entity.TeamMember)
				if currTeamMember.UserID == userID {
					teamIDs = append(teamIDs, currTeamMember.TeamID)
				}
			}

			return nil
		},
	})
	return teamIDs, err
}

func (t TeamMember) FindTeamMemberIDsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]uint64, *errs.Error) {
	var teamMemberIDs []uint64
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currTeamMember := rawRow.(entity.TeamMember)
				if currTeamMember.TeamID == teamID {
					teamMemberIDs = append(teamMemberIDs, currTeamMember.UserID)
				}
			}

			return nil
		},
	})
	return teamMemberIDs, err
}

func (t TeamMember) FindTeamMembersByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.TeamMember, *errs.Error) {
	var teamMembers []entity.TeamMember
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currTeamMember := rawRow.(entity.TeamMember)
				if currTeamMember.TeamID == teamID {
					teamMembers = append(teamMembers, currTeamMember)
				}
			}

			return nil
		},
	})
	return teamMembers, err
}

func (t TeamMember) FindTeamMemberWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64, userID uint64) (entity.TeamMember, *errs.Error) {
	var teamMember entity.TeamMember
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currTeamMember := rawRow.(entity.TeamMember)
				if currTeamMember.TeamID == teamID && currTeamMember.UserID == userID {
					teamMember = currTeamMember
					return nil
				}
			}

			return &errs.Error{
				Code:    errs.NotFound,
				Message: fmt.Sprintf("row not found: teamID=%v, userID=%v", teamID, userID),
			}
		},
	})
	return teamMember, err
}

func (t TeamMember) CreateTeamMember(ct context.Context, tx *transaction.Transaction, teamMember entity.TeamMember) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currTeamMember := row.(entity.TeamMember)
				if currTeamMember.TeamID == teamMember.TeamID &&
					currTeamMember.UserID == teamMember.UserID {
					return errs.NewError(errs.Unknown, fmt.Sprintf("row already exist: teamID=%v, userID=%v", teamMember.TeamID, teamMember.UserID))
				}
			}

			table.Rows = append(table.Rows, teamMember)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currTeamMember := row.(entity.TeamMember)
				if currTeamMember.TeamID == teamMember.TeamID &&
					currTeamMember.UserID == teamMember.UserID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (t TeamMember) UpdateTeamMember(ct context.Context, tx *transaction.Transaction, teamMember entity.TeamMember) *errs.Error {
	oldTeamMember, internalErr := t.FindTeamMemberWithTx(ct, tx, teamMember.TeamID, teamMember.UserID)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberTableName)
			if err != nil {
				return err
			}

			for i, row := range table.Rows {
				currTeamMember := row.(entity.TeamMember)
				if currTeamMember.TeamID == teamMember.TeamID &&
					currTeamMember.UserID == teamMember.UserID {
					table.Rows[i] = teamMember
					return nil
				}
			}

			return errs.NewError(errs.Unknown, fmt.Sprintf("row not exist: teamID=%v, userID=%v", teamMember.TeamID, teamMember.UserID))
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberTableName)
			if err != nil {
				return err
			}

			for index, row := range table.Rows {
				currTeamMember := row.(entity.TeamMember)
				if currTeamMember.TeamID == teamMember.TeamID &&
					currTeamMember.UserID == teamMember.UserID {
					table.Rows[index] = oldTeamMember
				}
			}

			return nil
		},
	})
}

func (t TeamMember) DeleteTeamMember(ct context.Context, tx *transaction.Transaction, teamID uint64, userID uint64) *errs.Error {
	oldTeamMember, internalErr := t.FindTeamMemberWithTx(ct, tx, teamID, userID)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currTeamMember := row.(entity.TeamMember)
				if currTeamMember.TeamID != teamID || currTeamMember.UserID != userID {
					rows = append(rows, currTeamMember)
				}
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberTableName)
			if err != nil {
				return err
			}

			table.Rows = append(table.Rows, oldTeamMember)
			return nil
		},
	})
}

func NewTeamMember(db *dbtest.InMemoryDB, transactionFactory transaction.Factory) TeamMember {
	return TeamMember{db: db, transactionFactory: transactionFactory}
}
