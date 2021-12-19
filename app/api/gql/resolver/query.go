package resolver

import (
	"context"
	"log"
	"sort"

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

func (q Query) Tasks(args struct{ Input *TaskFilter }) ([]Task, error) {
	tasks := q.deps.Data.FilterTasks(func(t entity.Task) bool {
		return taskFilterFunc(t, args.Input)
	})
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].ID < tasks[j].ID
	})
	return newTasks(q.deps, tasks), nil
}

func (q Query) Me(ctx context.Context) (User, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	user, err := q.deps.Data.GetUser(userID)
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
func (q Query) Teams(ctx context.Context, args struct {
	IDs *[]graphql.ID
}) ([]Team, error) {
	var teams []entity.Team

	if args.IDs == nil {
		teams = q.deps.Data.FilterTeams(func(team entity.Team) bool { return true })
	} else {
		idsMap, err := toIDsMap(*args.IDs)
		if err != nil {
			return nil, err
		}

		teams = q.deps.Data.FilterTeams(func(team entity.Team) bool {
			_, ok := idsMap[team.ID]
			return ok
		})
	}

	return newTeams(q.deps, teams), nil
}

func NewQuery(deps *Dependencies) Query {
	return Query{
		deps: deps,
	}
}
