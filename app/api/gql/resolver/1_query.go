package resolver

import (
	"context"
	"log"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Query struct {
	deps *Dependencies
}

func (q Query) Task(args struct {
	ID graphql.ID
}) (Task, error) {
	task, err := q.deps.Data.GetTask(args.ID)
	if err != nil {
		return Task{}, err
	}
	return newTask(q.deps, task), nil
}

func (q Query) Tasks() ([]Task, error) {
	tasks := q.deps.Data.FilterTasks(func(t entity.Task) bool { return true })
	return newTasks(q.deps, tasks), nil
}

func (q Query) Me(ctx context.Context) (User, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	user, err := q.deps.userRepo.FindUser(userID)
	if err != nil {
		log.Println(err)
		return User{}, err
	}
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	return newUser(q.deps, user), nil
}

func NewQuery(deps *Dependencies) Query {
	return Query{
		deps: deps,
	}
}
