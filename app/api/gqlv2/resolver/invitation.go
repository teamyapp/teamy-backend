package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type Invitation struct {
	deps       Dependencies
	invitation entityv2.Invitation
}

func (i Invitation) ID(ct context.Context) graphql.ID {
	return toGraphQLID(i.invitation.ID)
}

func (i Invitation) Sender(ct context.Context) (User, error) {
	sender, err := i.deps.userDao.FindUserByID(i.invitation.SenderUserID)
	if err != nil {
		return User{}, err
	}

	return User{
		user: sender,
		deps: i.deps,
	}, nil
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

	return &User{
		user: receiver,
		deps: i.deps,
	}, nil
}

func (i Invitation) JoiningTeam(ct context.Context) (Team, error) {
	team, err := i.deps.teamDao.FindTeamByID(i.invitation.TeamID)
	if err != nil {
		return Team{}, err
	}

	return Team{
		team: team,
		deps: i.deps,
	}, nil
}

func (i Invitation) ExpireAt(ct context.Context) graphql.Time {
	return toGraphQLTime(i.invitation.ExpireAt)
}

func (i Invitation) CreateAt(ct context.Context) graphql.Time {
	return toGraphQLTime(i.invitation.CreatedAt)
}

func (i Invitation) Status(ct context.Context) entityv2.InvitationStatus {
	return i.invitation.Status
}

func (i Invitation) Code(ct context.Context) string {
	return i.invitation.Code
}
