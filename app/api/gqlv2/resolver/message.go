package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type Message struct {
}

func (Message) ID(ctx context.Context) (graphql.ID, error) {
	panic("implement me")
}

func (Message) Body(ctx context.Context) (string, error) {
	panic("implement me")
}

func (Message) CreatedAt(ctx context.Context) (graphql.Time, error) {
	panic("implement me")
}

func (Message) UpdatedAt(ctx context.Context) (*graphql.Time, error) {
	panic("implement me")
}
