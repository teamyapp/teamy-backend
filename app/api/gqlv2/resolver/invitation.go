package resolver

import (
	"context"

	"github.com/teamyapp/teamy-backend/app/entityv2"

	"github.com/graph-gophers/graphql-go"
)

type Invitation struct {
	deps       Dependencies
	invitation entityv2.Invitation
}

func (i Invitation) ID(ctx context.Context) graphql.ID {
	return toGraphQLID(i.invitation.ID)
}

func (i Invitation) Sender(ctx context.Context) (User, error) {
	sender, err := i.deps.userDao.FindUserByID(i.invitation.SenderUserID)
	if err != nil {
		return User{}, err
	}

	return User{
		user: sender,
		deps: i.deps,
	}, nil
}

func (i Invitation) ReceiverFirstName(ctx context.Context) *string {
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

	receiver, err := i.deps.userDao.FindUserByID(*i.invitation.ReceiverUserID)
	if err != nil {
		return &User{}, err
	}

	return &User{
		user: receiver,
		deps: i.deps,
	}, nil
}

func (i Invitation) JoiningTeam(ctx context.Context) (Team, error) {
	team, err := i.deps.teamDao.FindTeamByID(i.invitation.TeamID)
	if err != nil {
		return Team{}, err
	}

	return Team{
		team: team,
		deps: i.deps,
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
