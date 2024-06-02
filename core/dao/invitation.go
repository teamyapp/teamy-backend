package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Invitation interface {
	FindInvitationByID(ct context.Context, invitationID uint64) (entity.Invitation, *errs.Error)
	FindInvitationsByIDsWithTx(ct context.Context, tx *transaction.Transaction, invitationIDs []uint64) ([]entity.Invitation, *errs.Error)
	FindInvitationsByTeamID(ct context.Context, teamID uint64) ([]entity.Invitation, *errs.Error)
	FindAllInvitations(ct context.Context) ([]entity.Invitation, *errs.Error)
	FindAllInvitationsWithTx(ct context.Context, tx *transaction.Transaction) ([]entity.Invitation, *errs.Error)
	FindInvitationByIDWithTx(ct context.Context, tx *transaction.Transaction, invitationID uint64) (entity.Invitation, *errs.Error)
	FindInvitationsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.Invitation, *errs.Error)
	CreateInvitation(ct context.Context, tx *transaction.Transaction, invitation entity.Invitation) *errs.Error
	UpdateInvitation(ct context.Context, tx *transaction.Transaction, invitation entity.Invitation) *errs.Error
	DeleteInvitation(ct context.Context, tx *transaction.Transaction, invitationID uint64) *errs.Error
}
