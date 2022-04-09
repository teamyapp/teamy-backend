package resolver

import (
	"context"

	"github.com/teamyapp/teamy-backend/app/dao"

	"github.com/graph-gophers/graphql-go"
)

type Team struct {
	team dao.Team
	deps Dependencies
}

func (Team) ID(ctx context.Context) (graphql.ID, error) {
	panic("implement me")
}

func (Team) Name(ctx context.Context) (string, error) {
	panic("implement me")
}

func (Team) IconURL(ctx context.Context) (*string, error) {
	panic("implement me")
}

func (Team) CreatedAt(ctx context.Context) (graphql.Time, error) {
	panic("implement me")
}

func (Team) Creator(ctx context.Context) (User, error) {
	panic("implement me")
}

func (Team) Owner(ctx context.Context) (User, error) {
	panic("implement me")
}

func (Team) Members(ctx context.Context) ([]User, error) {
	panic("implement me")
}

func (Team) Tasks(ctx context.Context, args struct {
	Filter TaskFilter
}) ([]Task, error) {
	panic("implement me")
}

func (Team) Invitations(ctx context.Context) ([]Invitation, error) {
	panic("implement me")
}
