package resolver

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Invitation struct {
	deps          *Dependencies
	Entity
	invitation entity.Invitation
	query      Query
}

func (i Invitation) Inviter() User {
	panic("need implementation")
}

func (i Invitation) NewMember() *User {
	panic("need implementation")
}

func (i Invitation) NewMemberEmail() *string {
	panic("need implementation")
}

func (i Invitation) Team() Team {
	panic("need implementation")
}

func (i Invitation) Expiration() graphql.Time {
	expiration := toGraphQLTime(&i.invitation.Expiration)
	return *expiration
}

func (i Invitation) Status() InvitationStatus {
	return toGraphQLInvitationStatus(i.invitation.Status)
}

func newInvitation(deps *Dependencies, invitation entity.Invitation) Invitation {
	return Invitation{
		Entity:        Entity{entity: invitation.Entity},
		deps:          deps,
		invitation:    invitation,
	}
}
