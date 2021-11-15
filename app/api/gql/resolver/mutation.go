package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type Mutation struct {
}

func (m Mutation) CreateTask(ctx context.Context, args struct {
	TaskID graphql.ID
	Task   TaskInput
}) graphql.ID {
	panic("not implemented")
}

func (m Mutation) DeleteTask(ctx context.Context, args struct {
	TaskID graphql.ID
}) bool {
	panic("not implemented")
}

func (m Mutation) UpdateTask(ctx context.Context, args struct {
	TaskID graphql.ID
	Task   TaskInput
}) bool {
	panic("not implemented")
}

func (m Mutation) PerformTaskAction(ctx context.Context, args struct {
	TaskID graphql.ID
	Action TaskAction
}) bool {
	panic("not implemented")
}
