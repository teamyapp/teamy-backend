package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Invitation struct {
}

func (Invitation) ID(ctx context.Context) (graphql.ID, error) {
}

func (Invitation) Sender(ctx context.Context) (User, error) {
}

func (Invitation) ReceiverFirstName(ctx context.Context) (string, error) {
}

func (Invitation) ReceiverLastName(ctx context.Context) (*string, error) {
}

func (Invitation) ReceiverEmail(ctx context.Context) (*string, error) {
}

func (Invitation) Receiver(ctx context.Context) (*User, error) {
}

func (Invitation) JoiningTeam(ctx context.Context) (Team, error) {
}

func (Invitation) ExpireAt(ctx context.Context) (graphql.Time, error) {
}

func (Invitation) CreateAt(ctx context.Context) (graphql.Time, error) {
}

func (Invitation) Status(ctx context.Context) (entity.InvitationStatus, error) {
}

func (Invitation) Code(ctx context.Context) (*string, error) {
}
