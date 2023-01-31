package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

func (m Mutation) CreateApp(ct context.Context, args struct {
	Name string
}) (App, error) {
	panic("implement me")
}

func (m Mutation) UpdateApp(ct context.Context, args struct {
	AppID graphql.ID
	Input struct {
		AppName             *string
		ActiveVersionNumber *int32
		Description         *string
	}
}) (App, error) {
	panic("implement me")
}

func (m Mutation) RefreshAppSecret(ct context.Context, args struct {
	AppID graphql.ID
}) (App, error) {
	panic("implement me")
}

func (m Mutation) DeleteApp(ct context.Context, args struct {
	AppID graphql.ID
}) (App, error) {
	panic("implement me")
}
