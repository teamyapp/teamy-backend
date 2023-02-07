package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Invitation interface {
	FindAllInvitations(ct context.Context) ([]entity.Invitation, *errs.Error)
	FindInvitationByID(ct context.Context, invitationID uint64) (entity.Invitation, *errs.Error)
	FindInvitationsByTeamID(ct context.Context, teamID uint64) ([]entity.Invitation, *errs.Error)
	CreateInvitation(ct context.Context, invitation entity.Invitation) *errs.Error
	UpdateInvitation(ct context.Context, invitation entity.Invitation) *errs.Error
	DeleteInvitation(ct context.Context, invitationID uint64) *errs.Error
}
