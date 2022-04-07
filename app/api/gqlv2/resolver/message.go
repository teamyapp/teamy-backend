package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type Message struct {
}

func (Message) ID(ctx context.Context) (graphql.ID, error) {
}

func (Message) Body(ctx context.Context) (string, error) {
}

func (Message) CreatedAt(ctx context.Context) (graphql.Time, error) {
}

func (Message) UpdatedAt(ctx context.Context) (*graphql.Time, error) {
}
