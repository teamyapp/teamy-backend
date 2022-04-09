package resolver

import (
	"context"

	"github.com/teamyapp/teamy-backend/app/entityv2"

	"github.com/graph-gophers/graphql-go"
)

type Invitation struct {
	invitation entityv2.Invitation
	deps       *Dependencies
}

func (i Invitation) ID(ctx context.Context) (graphql.ID, error) {
	return toGraphQLID(i.invitation.ID), nil
}

func (i Invitation) Sender(ctx context.Context) (User, error) {
	sender, err := i.deps.userDao.FindUser(i.invitation.SenderUserID)
	if err != nil {
		return User{}, err
	}

	return User{
		sender,
		i.deps,
	}, nil
}

func (i Invitation) ReceiverFirstName(ctx context.Context) string {
	return i.invitation.ReceiverFirstName
}

func (i Invitation) ReceiverLastName(ctx context.Context) *string {
	return i.invitation.ReceiverLastName
}

func (i Invitation) ReceiverEmail(ctx context.Context) *string {
	return i.invitation.ReceiverEmail
}

func (i Invitation) Receiver(ctx context.Context) (*User, error) {
	if i.invitation.ReceiverUserID == nil {
		return nil, nil
	}

	receiver, err := i.deps.userDao.FindUser(*i.invitation.ReceiverUserID)
	if err != nil {
		return &User{}, err
	}

	return &User{
		receiver,
		i.deps,
	}, nil
}

func (i Invitation) JoiningTeam(ctx context.Context) (Team, error) {
	team, err := i.deps.teamDao.FindTeam(i.invitation.TeamID)
	if err != nil {
		return Team{}, err
	}

	return Team{
		team,
		i.deps,
	}, nil
}

func (i Invitation) ExpireAt(ctx context.Context) graphql.Time {
	return toGraphQLTime(i.invitation.ExpireAt)
}

func (i Invitation) CreateAt(ctx context.Context) graphql.Time {
	return toGraphQLTime(i.invitation.CreatedAt)
}

func (i Invitation) Status(ctx context.Context) entityv2.InvitationStatus {
	return i.invitation.Status
}

func (i Invitation) Code(ctx context.Context) string {
	return i.invitation.Code
}
