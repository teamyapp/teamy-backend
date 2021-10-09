package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type Query struct {
}

func (q Query) ExecutionMode(ctx context.Context) ExecutionMode {
	panic("not implemented")
}

func (q Query) Task(ctx context.Context, args struct {
	TaskID graphql.ID
}) *Task {
	panic("not implemented")
}
