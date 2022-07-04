package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Invitation struct {
	deps       *Dependencies
	invitation entity.Invitation
}

func (i Invitation) ID(ct context.Context) graphql.ID {
	return toGraphQLID(i.invitation.ID)
}

func (i Invitation) Sender(ct context.Context) (User, error) {
	sender, err := i.deps.userDao.FindUserByID(i.invitation.SenderUserID)
	if err != nil {
		return User{}, err
	}

	return newUser(i.deps, sender), nil
}

func (i Invitation) ReceiverFirstName(ct context.Context) *string {
	return i.invitation.ReceiverFirstName
}

func (i Invitation) ReceiverLastName(ct context.Context) *string {
	return i.invitation.ReceiverLastName
}

func (i Invitation) ReceiverEmail(ct context.Context) *string {
	return i.invitation.ReceiverEmail
}

func (i Invitation) Receiver(ct context.Context) (*User, error) {
	if i.invitation.ReceiverUserID == nil {
		return nil, nil
	}

	receiver, err := i.deps.userDao.FindUserByID(*i.invitation.ReceiverUserID)
	if err != nil {
		return &User{}, err
	}

	gqlUser := newUser(i.deps, receiver)
	return &gqlUser, nil
}

func (i Invitation) JoiningTeam(ct context.Context) (Team, error) {
	team, err := i.deps.teamDao.FindTeamByID(i.invitation.TeamID)
	if err != nil {
		return Team{}, err
	}

	return newTeam(i.deps, team), nil
}

func (i Invitation) ExpireAt(ct context.Context) graphql.Time {
	return toGraphQLTime(i.invitation.ExpireAt)
}

func (i Invitation) CreatedAt(ct context.Context) graphql.Time {
	return toGraphQLTime(i.invitation.CreatedAt)
}

func (i Invitation) UpdatedAt(ct context.Context) *graphql.Time {
	return toGraphQLTimePtr(i.invitation.UpdatedAt)
}

func (i Invitation) Status(ct context.Context) entity.InvitationStatus {
	return i.invitation.Status
}

func (i Invitation) Code(ct context.Context) string {
	return i.invitation.Code
}

func newInvitation(deps *Dependencies, invitation entity.Invitation) Invitation {
	return Invitation{
		deps:       deps,
		invitation: invitation,
	}
}
