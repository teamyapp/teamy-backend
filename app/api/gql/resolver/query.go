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
	// TODO: (Begin) remove once JSON data feed is ready
	taskID, err := fromGraphQLID(args.ID)
	if err != nil {
		return Task{}, err
	}
	task, err := q.deps.taskRepo.FindTaskByID(taskID)
	// TODO: (End) remove once JSON data feed is ready

	// TODO: (Begin) enable once JSON data feed is ready
	//task, err := q.deps.Data.GetTask(args.ID)
	// TODO: (End) enable once JSON data feed is ready
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

	// TODO: (Begin) remove once JSON data feed is ready
	user, err := q.deps.userRepo.FindUser(userID)
	// TODO: (End) remove once JSON data feed is ready

	// TODO: (Begin) enable once JSON data feed is ready
	// user, err := q.deps.Data.GetUser(userID)
	// TODO: (End) enable once JSON data feed is ready
	if err != nil {
		log.Printf("%+v\n", err)
		return User{}, err
	}
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	return newUser(q.deps, user), nil
}

// debug only
func (q Query) Teams(ctx context.Context) ([]Team, error) {
	teams := q.deps.Data.FilterTeams(func(t entity.Team) bool { return true })
	return newTeams(q.deps, teams), nil
}

func NewQuery(deps *Dependencies) Query {
	return Query{
		deps: deps,
	}
}
