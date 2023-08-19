package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Group interface {
	ID(ctx context.Context) graphql.ID
	Type(ctx context.Context) entity.GroupType
	Name(ctx context.Context) string
	CreatedAt(ctx context.Context) graphql.Time
	UpdatedAt(ctx context.Context) *graphql.Time
}

type FilterGroup interface {
	ID(ctx context.Context) graphql.ID
	Type(ctx context.Context) entity.GroupType
	Name(ctx context.Context) string
	Filter(ctx context.Context) string
	CreatedAt(ctx context.Context) graphql.Time
	UpdatedAt(ctx context.Context) *graphql.Time
}

func (m Mutation) DeleteGroup(
	ctx context.Context,
	args struct {
		GroupID graphql.ID
	},
) Group {
	panic("implement me")
}
