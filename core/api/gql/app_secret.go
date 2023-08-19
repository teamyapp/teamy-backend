package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type AppSecret struct {
}

func (a AppSecret) ID(ctx context.Context) graphql.ID {
	panic("not implemented")
}

func (a AppSecret) Name(ctx context.Context) string {
	panic("not implemented")
}

func (a AppSecret) AddedAt(ctx context.Context) graphql.Time {
	panic("not implemented")
}

func (a AppSecret) AddedBy(ctx context.Context) User {
	panic("not implemented")
}

func (a AppSecret) LastUsedAt(ctx context.Context) *graphql.Time {
	panic("not implemented")
}

func (a AppSecret) App(ctx context.Context) App {
	panic("not implemented")
}

func (m Mutation) CreateAppSecret(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			Name string
		}
	}) AppSecret {
	panic("not implemented")
}

func (m Mutation) UpdateAppSecret(
	ctx context.Context,
	args struct {
		SecretID graphql.ID
		Input    struct {
			Name string
		}
	}) AppSecret {
	panic("not implemented")
}

func (m Mutation) DeleteAppSecret(
	ctx context.Context,
	args struct {
		SecretID graphql.ID
	},
) AppSecret {
	panic("not implemented")
}
