package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Resolver struct {
	Query
	Mutation
}

type DataStore interface {
	taskStore
}

type taskStore interface {
	CreateTask(c context.Context, task TaskInput, creatorID graphql.ID) (entity.Task, error)
}

func NewResolver(deps *Dependencies) Resolver {
	query := NewQuery(deps)
	return Resolver{
		Query:    NewQuery(deps),
		Mutation: NewMutation(deps, &query),
	}
}
