package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Invitation struct {
}

func (Invitation) ID(ctx context.Context) (graphql.ID, error) {
	panic("implement me")
}

func (Invitation) Sender(ctx context.Context) (User, error) {
	panic("implement me")
}

func (Invitation) ReceiverFirstName(ctx context.Context) (string, error) {
	panic("implement me")
}

func (Invitation) ReceiverLastName(ctx context.Context) (*string, error) {
	panic("implement me")
}

func (Invitation) ReceiverEmail(ctx context.Context) (*string, error) {
	panic("implement me")
}

func (Invitation) Receiver(ctx context.Context) (*User, error) {
	panic("implement me")
}

func (Invitation) JoiningTeam(ctx context.Context) (Team, error) {
	panic("implement me")
}

func (Invitation) ExpireAt(ctx context.Context) (graphql.Time, error) {
	panic("implement me")
}

func (Invitation) CreateAt(ctx context.Context) (graphql.Time, error) {
	panic("implement me")
}

func (Invitation) Status(ctx context.Context) (entity.InvitationStatus, error) {
	panic("implement me")
}

func (Invitation) Code(ctx context.Context) (*string, error) {
	panic("implement me")
}
