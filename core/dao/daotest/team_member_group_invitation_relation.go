package daotest

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamMemberGroupInvitationRelation struct {
	db                 *dbtest.InMemoryDB
	transactionFactory transaction.Factory
}

var _ dao.TeamMemberGroupInvitationRelation = (*TeamMemberGroupInvitationRelation)(nil)

func (t TeamMemberGroupInvitationRelation) FindInvitationIDsByTeamMemberGroupID(
	ct context.Context,
	tx *transaction.Transaction,
	teamMemberGroupID uint64,
) ([]uint64, *errs.Error) {
	var invitationIDs []uint64
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupInvitationRelationTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currRelation := rawRow.(entity.TeamMemberGroupInvitationRelation)
				if currRelation.GroupID == teamMemberGroupID {
					invitationIDs = append(invitationIDs, currRelation.InvitationID)
				}
			}

			return nil
		},
	})
	return invitationIDs, err
}

func (t TeamMemberGroupInvitationRelation) CreateTeamMemberGroupInvitationRelation(
	ct context.Context,
	tx *transaction.Transaction,
	relation entity.TeamMemberGroupInvitationRelation,
) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupInvitationRelationTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currRelation := row.(entity.TeamMemberGroupInvitationRelation)
				if currRelation.GroupID == relation.GroupID && currRelation.InvitationID == relation.InvitationID {
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
			table, err := t.db.GetTable(TeamMemberGroupInvitationRelationTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currRelation := row.(entity.TeamMemberGroupInvitationRelation)
				if currRelation.GroupID == relation.GroupID && currRelation.InvitationID == relation.InvitationID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})

}

func (t TeamMemberGroupInvitationRelation) DeleteTeamMemberGroupInvitationRelation(
	ct context.Context,
	tx *transaction.Transaction,
	relation entity.TeamMemberGroupInvitationRelation,
) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupInvitationRelationTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currRelation := row.(entity.TeamMemberGroupInvitationRelation)
				if currRelation.GroupID == relation.GroupID && currRelation.InvitationID == relation.InvitationID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupInvitationRelationTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currRelation := row.(entity.TeamMemberGroupInvitationRelation)
				if currRelation.GroupID == relation.GroupID && currRelation.InvitationID == relation.InvitationID {
					return errs.NewError(
						errs.Unknown,
						"row already exist",
					)
				}
			}

			table.Rows = append(table.Rows, relation)
			return nil
		},
	})
}

func (t TeamMemberGroupInvitationRelation) DeleteTeamMemberGroupInvitationRelationsByGroupID(
	ct context.Context,
	tx *transaction.Transaction,
	groupID uint64,
) *errs.Error {
	deletedTeamMemberGroupInvitationRelations := make([]entity.TeamMemberGroupInvitationRelation, 0)
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupInvitationRelationTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currRelation := row.(entity.TeamMemberGroupInvitationRelation)
				if currRelation.GroupID == groupID {
					deletedTeamMemberGroupInvitationRelations = append(deletedTeamMemberGroupInvitationRelations, currRelation)
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TeamMemberGroupInvitationRelationTableName)
			if err != nil {
				return err
			}

			for _, relation := range deletedTeamMemberGroupInvitationRelations {
				table.Rows = append(table.Rows, relation)
			}

			return nil
		},
	})
}

func NewTeamMemberGroupInvitationRelation(db *dbtest.InMemoryDB, transactionFactory transaction.Factory) TeamMemberGroupInvitationRelation {
	return TeamMemberGroupInvitationRelation{
		db:                 db,
		transactionFactory: transactionFactory,
	}
}
