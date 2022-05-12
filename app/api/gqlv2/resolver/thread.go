package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type Thread struct {
}

func (Thread) ID(ct context.Context) (graphql.ID, error) {
	panic("implement me")
}

func (Thread) Messages(ct context.Context) ([]Message, error) {
	panic("implement me")
}
