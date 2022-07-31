package dao

import (
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Invitation interface {
	FindAllInvitations() ([]entity.Invitation, error)
	FindInvitationByID(invitationID uint64) (entity.Invitation, error)
	FindInvitationsByTeamID(teamID uint64) ([]entity.Invitation, error)
	CreateInvitation(invitation entity.Invitation) error
	UpdateInvitation(invitation entity.Invitation) error
	DeleteInvitation(invitationID uint64) error
}
