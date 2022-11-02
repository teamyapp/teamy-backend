package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

func (m Mutation) CreateApp(ct context.Context) (App, error) {
	panic("implement me")
}

func (m Mutation) UpdateApp(ct context.Context, args struct {
	Input struct {
		ActiveVersionNumber *int32
	}
}) (App, error) {
	panic("implement me")
}

func (m Mutation) RefreshAppSecret(ct context.Context) (App, error) {
	panic("implement me")
}

func (m Mutation) DeleteApp(ct context.Context, args struct {
	AppID graphql.ID
}) (App, error) {
	panic("implement me")
}
