package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type Thread struct {
}

func (Thread) ID(ctx context.Context) (graphql.ID, error) {
}

func (Thread) Messages(ctx context.Context) ([]Message, error) {
}
