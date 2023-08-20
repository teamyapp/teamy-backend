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

type Invitation struct {
	db                 *dbtest.InMemoryDB
	transactionFactory transaction.Factory
}

var _ dao.Invitation = (*Invitation)(nil)

func (i Invitation) FindInvitationByID(ct context.Context, invitationID uint64) (entity.Invitation, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := i.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.Invitation{}, err
	}

	defer tx.Rollback()
	return i.FindInvitationByIDWithTx(ct, tx, invitationID)
}

func (i Invitation) FindInvitationsByTeamID(ct context.Context, teamID uint64) ([]entity.Invitation, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := i.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return i.FindInvitationsByTeamIDWithTx(ct, tx, teamID)
}

func (i Invitation) FindAllInvitations(ct context.Context) ([]entity.Invitation, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := i.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return i.FindAllInvitationsWithTx(ct, tx)
}

func (i Invitation) FindAllInvitationsWithTx(ct context.Context, tx *transaction.Transaction) ([]entity.Invitation, *errs.Error) {
	var invitations []entity.Invitation
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := i.db.GetTable(InvitationTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currInvitation := rawRow.(entity.Invitation)
				invitations = append(invitations, currInvitation)
			}

			return nil
		},
	})
	return invitations, err
}

func (i Invitation) FindInvitationByIDWithTx(ct context.Context, tx *transaction.Transaction, invitationID uint64) (entity.Invitation, *errs.Error) {
	var invitation entity.Invitation
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := i.db.GetTable(InvitationTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currInvitation := rawRow.(entity.Invitation)
				if currInvitation.ID == invitationID {
					invitation = currInvitation
					return nil
				}
			}

			return errs.NewError(errs.NotFound, fmt.Sprintf("row not found: invitationID=%v", invitationID))
		},
	})
	return invitation, err
}

func (i Invitation) FindInvitationsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.Invitation, *errs.Error) {
	var invitations []entity.Invitation
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := i.db.GetTable(InvitationTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currInvitation := rawRow.(entity.Invitation)
				if currInvitation.TeamID == teamID {
					invitations = append(invitations, currInvitation)
				}
			}

			return nil
		},
	})
	return invitations, err
}

func (i Invitation) CreateInvitation(ct context.Context, tx *transaction.Transaction, invitation entity.Invitation) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := i.db.GetTable(InvitationTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currInvitation := row.(entity.Invitation)
				if currInvitation.ID == invitation.ID {
					return errs.NewError(errs.Unknown, fmt.Sprintf("row already exist: invitationID=%v", invitation.ID))
				}
			}

			table.Rows = append(table.Rows, invitation)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := i.db.GetTable(InvitationTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currInvitation := row.(entity.Invitation)
				if currInvitation.ID == invitation.ID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (i Invitation) UpdateInvitation(ct context.Context, tx *transaction.Transaction, invitation entity.Invitation) *errs.Error {
	oldInvitation, internalErr := i.FindInvitationByIDWithTx(ct, tx, invitation.ID)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := i.db.GetTable(InvitationTableName)
			if err != nil {
				return err
			}

			for i, row := range table.Rows {
				currInvitation := row.(entity.Invitation)
				if currInvitation.ID == invitation.ID {
					table.Rows[i] = invitation
					return nil
				}
			}

			return errs.NewError(errs.Unknown, fmt.Sprintf("row not exist: invitationID=%v", invitation.ID))
		},
		Undo: func() *errs.Error {
			table, err := i.db.GetTable(InvitationTableName)
			if err != nil {
				return err
			}

			for index, row := range table.Rows {
				currInvitation := row.(entity.Invitation)
				if currInvitation.ID == invitation.ID {
					table.Rows[index] = oldInvitation
				}
			}

			return nil
		},
	})
}

func (i Invitation) DeleteInvitation(ct context.Context, tx *transaction.Transaction, invitationID uint64) *errs.Error {
	oldInvitation, internalErr := i.FindInvitationByIDWithTx(ct, tx, invitationID)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := i.db.GetTable(InvitationTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currInvitation := row.(entity.Invitation)
				if currInvitation.ID != invitationID {
					rows = append(rows, currInvitation)
				}
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := i.db.GetTable(InvitationTableName)
			if err != nil {
				return err
			}

			table.Rows = append(table.Rows, oldInvitation)
			return nil
		},
	})
}

func NewInvitation(db *dbtest.InMemoryDB, transactionFactory transaction.Factory) Invitation {
	return Invitation{db: db, transactionFactory: transactionFactory}
}
