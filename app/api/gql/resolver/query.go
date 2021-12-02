package resolver

import (
	"context"
	"log"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/api/gqlv2/resolver"
)

type Query struct {
	deps          *Dependencies
	prototypeDeps *resolver.Dependencies
}

func (q Query) ExecutionMode(ctx context.Context) (ExecutionMode, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return ExecutionMode{}, err
	}
	return newExecutionMode(q.deps, q.prototypeDeps, userID), nil
}

func (q Query) Task(args struct {
	ID graphql.ID
}) (Task, error) {
	id, err := strconv.Atoi(string(args.ID))
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	task, err := q.deps.taskRepo.FindTaskByID(oneEntity.ID(id))
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	return newTask(q.deps, q.prototypeDeps, task), nil
}

func (q Query) Me(ctx context.Context) (User, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	user, err := q.deps.userRepo.GetUser(userID)
	if err != nil {
		log.Println(err)
		return User{}, err
	}
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	return newUser(q.deps, q.prototypeDeps, user), nil
}

func NewQuery(deps *Dependencies, prototypeDeps *resolver.Dependencies) Query {
	return Query{
		deps:          deps,
		prototypeDeps: prototypeDeps,
	}
}
