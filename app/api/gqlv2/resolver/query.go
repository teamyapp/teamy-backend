package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

func (r Root) Tasks(c context.Context, args struct{ ID graphql.ID }) ([]Task, error) {
	tasks := r.Deps.Data.GetTasks([]graphql.ID{args.ID})
	for i := range tasks {
		tasks[i].deps = r.Deps
	}
	return tasks, nil
}

func (r Root) Me() (User, error) {
	u, err := r.Deps.Data.GetUser("1")
	u.deps = r.Deps
	return u, err
}
