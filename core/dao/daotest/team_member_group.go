package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/entity"
)

type TeamMemberGroup struct {
	db                 *dbtest.InMemoryDB
	transactionFactory transaction.Factory
}

var _ dao.TeamMemberGroup = (*TeamMemberGroup)(nil)

func (t TeamMemberGroup) FindMemberGroupByID(ct context.Context, tx *transaction.Transaction, id uint64) (entity.TeamMemberGroup, *errs.Error) {
	var teamMemberGroup entity.TeamMemberGroup
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currGroup := rawRow.(entity.TeamMemberGroup)
				if currGroup.ID == id {
					teamMemberGroup = currGroup
					return nil
				}
			}

			return &errs.Error{
				Code:    errs.NotFound,
				Message: fmt.Sprintf("row not found: id=%v", id),
			}
		},
	})
	return teamMemberGroup, err
}

func (t TeamMemberGroup) FindMemberGroupsByTeamID(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.TeamMemberGroup, *errs.Error) {
	var teamMemberGroups []entity.TeamMemberGroup
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currGroup := rawRow.(entity.TeamMemberGroup)
				if currGroup.TeamID == teamID {
					teamMemberGroups = append(teamMemberGroups, currGroup)
				}
			}

			return nil
		},
	})
	return teamMemberGroups, err
}

func (t TeamMemberGroup) CreateMemberGroup(ct context.Context, tx *transaction.Transaction, memberGroup entity.TeamMemberGroup) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				teamMemberGroup := row.(entity.TeamMemberGroup)
				if teamMemberGroup.ID == memberGroup.ID {
					return errs.NewError(
						errs.Unknown,
						fmt.Sprintf("row already exist: id=%v",
							teamMemberGroup.ID,
						))
				}
			}

			table.Rows = append(table.Rows, memberGroup)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				teamMemberGroup := row.(entity.TeamMemberGroup)
				if teamMemberGroup.ID == memberGroup.ID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (t TeamMemberGroup) UpdateMemberGroup(ct context.Context, tx *transaction.Transaction, memberGroup entity.TeamMemberGroup) *errs.Error {
	oldTeamMemberGroup, internalErr := t.FindMemberGroupByID(ct, tx, memberGroup.ID)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupTableName)
			if err != nil {
				return err
			}

			for i, row := range table.Rows {
				currTeamMemberGroup := row.(entity.TeamMemberGroup)
				if currTeamMemberGroup.ID == memberGroup.ID {
					table.Rows[i] = memberGroup
					return nil
				}
			}

			return errs.NewError(errs.Unknown, fmt.Sprintf("row not exist: id=%v", memberGroup.ID))
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupTableName)
			if err != nil {
				return err
			}

			for index, row := range table.Rows {
				currTeamMemberGroup := row.(entity.TeamMemberGroup)
				if currTeamMemberGroup.ID == memberGroup.ID {
					table.Rows[index] = oldTeamMemberGroup
				}
			}

			return nil
		},
	})
}

func (t TeamMemberGroup) DeleteMemberGroup(ct context.Context, tx *transaction.Transaction, id uint64) *errs.Error {
	oldGroup, internalErr := t.FindMemberGroupByID(ct, tx, id)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currGroup := row.(entity.TeamMemberGroup)
				if currGroup.ID != id {
					rows = append(rows, currGroup)
				}
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupTableName)
			if err != nil {
				return err
			}

			table.Rows = append(table.Rows, oldGroup)
			return nil
		},
	})
}

func NewTeamMemberGroup(
	db *dbtest.InMemoryDB,
	transactionFactory transaction.Factory,
) TeamMemberGroup {
	return TeamMemberGroup{
		db:                 db,
		transactionFactory: transactionFactory,
	}
}
