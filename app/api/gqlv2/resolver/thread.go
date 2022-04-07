package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type Thread struct {
}

func (Thread) ID(ctx context.Context) (graphql.ID, error) {
	panic("implement me")
}

func (Thread) Messages(ctx context.Context) ([]Message, error) {
	panic("implement me")
}
