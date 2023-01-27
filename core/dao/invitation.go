package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/core/entity"
)

type Invitation interface {
	FindAllInvitations(ct context.Context) ([]entity.Invitation, error)
	FindInvitationByID(ct context.Context, invitationID uint64) (entity.Invitation, error)
	FindInvitationsByTeamID(ct context.Context, teamID uint64) ([]entity.Invitation, error)
	CreateInvitation(ct context.Context, invitation entity.Invitation) error
	UpdateInvitation(ct context.Context, invitation entity.Invitation) error
	DeleteInvitation(ct context.Context, invitationID uint64) error
}
