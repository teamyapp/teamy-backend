package dao

import (
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type Invitation interface {
	FindAllInvitations() ([]entityv2.Invitation, error)
	FindInvitationByID(id uint64) (entityv2.Invitation, error)
	FindInvitationsByTeamID(teamID uint64) ([]entityv2.Invitation, error)
	CreateInvitation(invitation entityv2.Invitation) error
	UpdateInvitation(invitation entityv2.Invitation) error
	DeleteInvitation(id uint64) error
}
