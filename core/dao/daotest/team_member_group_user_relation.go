package daotest

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/entity"
)

type TeamMemberGroupUserRelation struct {
	db                 *dbtest.InMemoryDB
	transactionFactory transaction.Factory
}

var _ dao.TeamMemberGroupUserRelation = (*TeamMemberGroupUserRelation)(nil)

func (t TeamMemberGroupUserRelation) FindMemberGroupUserIDsByMemberGroupID(ct context.Context, tx *transaction.Transaction, memberGroupID uint64) ([]uint64, *errs.Error) {
	var groupUserIDs []uint64
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupUserRelationTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currRelation := rawRow.(entity.TeamMemberGroupUserRelation)
				if currRelation.UserID == memberGroupID {
					groupUserIDs = append(groupUserIDs, currRelation.UserID)
				}
			}

			return nil
		},
	})
	return groupUserIDs, err
}

func (t TeamMemberGroupUserRelation) FindMemberGroupIDsByUserID(ct context.Context, tx *transaction.Transaction, userID uint64) ([]uint64, *errs.Error) {
	var memberGroupIDs []uint64
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currGroup := rawRow.(entity.TeamMemberGroup)
				if currGroup.ID == userID {
					memberGroupIDs = append(memberGroupIDs, currGroup.ID)
				}
			}

			return nil
		},
	})
	return memberGroupIDs, err
}

func (t TeamMemberGroupUserRelation) CreateMemberGroupUserRelation(ct context.Context, tx *transaction.Transaction, relation entity.TeamMemberGroupUserRelation) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupUserRelationTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currRelation := row.(entity.TeamMemberGroupUserRelation)
				if currRelation.GroupID == relation.GroupID && currRelation.UserID == relation.UserID {
					return errs.NewError(
						errs.Unknown,
						"row already exist",
					)
				}
			}

			table.Rows = append(table.Rows, relation)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupUserRelationTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currRelation := row.(entity.TeamMemberGroupUserRelation)
				if currRelation.GroupID == relation.GroupID && currRelation.UserID == relation.UserID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (t TeamMemberGroupUserRelation) DeleteMemberGroupUserRelation(ct context.Context, tx *transaction.Transaction, relation entity.TeamMemberGroupUserRelation) *errs.Error {
	var teamMemberGroup *entity.TeamMemberGroupUserRelation
	table, err := t.db.GetTable(TeamMemberGroupUserRelationTableName)
	if err != nil {
		return err
	}

	for _, rawRow := range table.Rows {
		groupMemberUserRelation := rawRow.(entity.TeamMemberGroupUserRelation)
		if groupMemberUserRelation.GroupID == relation.GroupID &&
			groupMemberUserRelation.UserID == relation.UserID {
			teamMemberGroup = &groupMemberUserRelation
			break
		}
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			rows := make([]interface{}, 0)
			for _, rawRow := range table.Rows {
				currGroupMemberUserRelation := rawRow.(entity.TeamMemberGroupUserRelation)
				if currGroupMemberUserRelation.GroupID != relation.GroupID ||
					currGroupMemberUserRelation.UserID != relation.UserID {
					rows = append(rows, currGroupMemberUserRelation)
				}
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			if teamMemberGroup != nil {
				table.Rows = append(table.Rows, *teamMemberGroup)
			}

			return nil
		},
	})
}

func NewTeamMemberGroupUserRelation(
	db *dbtest.InMemoryDB,
	transactionFactory transaction.Factory,
) *TeamMemberGroupUserRelation {
	return &TeamMemberGroupUserRelation{
		db:                 db,
		transactionFactory: transactionFactory,
	}
}
