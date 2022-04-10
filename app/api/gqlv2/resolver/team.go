package resolver

import (
	"context"

	"github.com/teamyapp/teamy-backend/app/entityv2"

	"github.com/graph-gophers/graphql-go"
)

type Team struct {
	deps Dependencies
	team entityv2.Team
}

func (t Team) ID(ctx context.Context) graphql.ID {
	return toGraphQLID(t.team.ID)
}

func (t Team) Name(ctx context.Context) (string, error) {
	panic("implement me")
}

func (t Team) IconURL(ctx context.Context) (*string, error) {
	panic("implement me")
}

func (t Team) CreatedAt(ctx context.Context) (graphql.Time, error) {
	panic("implement me")
}

func (t Team) Creator(ctx context.Context) (User, error) {
	panic("implement me")
}

func (t Team) Owner(ctx context.Context) (User, error) {
	panic("implement me")
}

func (t Team) Members(ctx context.Context) ([]User, error) {
	panic("implement me")
}

func (t Team) Tasks(ctx context.Context, args struct {
	Filter TaskFilter
}) ([]Task, error) {
	panic("implement me")
}

func (t Team) Invitations(ctx context.Context) ([]Invitation, error) {
	panic("implement me")
}
