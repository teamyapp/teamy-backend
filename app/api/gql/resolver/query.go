package resolver

import (
	"context"
	"log"
	"strconv"

	"github.com/graph-gophers/graphql-go"
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/one/identity"
	"github.com/teamyapp/teamy-backend/app/service"
)

type Query struct {
	taskService      service.Task
	executionService service.Execution
	userService service.User
}

func (q Query) ExecutionMode(ctx context.Context) (ExecutionMode, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return ExecutionMode{}, err
	}
	return ExecutionMode{
		userID:           userID,
		executionService: q.executionService,
	}, nil
}

func (q Query) Task(args struct {
	ID graphql.ID
}) (Task, error) {
	id, err := strconv.Atoi(string(args.ID))
	if err != nil {
		log.Println(err)
		return Task{}, err
	}
	ts, err := q.taskService.FindTask(oneEntity.ID(id))
	return newTask(ts), err
}

func (q Query) ActiveTeam(ctx context.Context) (*Team, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	team, err := q.executionService.GetActiveTeam(userID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if team == nil {
		return nil, nil
	}

	gqlTeam := newTeam(*team, q.userService)
	return &gqlTeam, nil
}

func NewQuery(
	taskService service.Task,
	executionService service.Execution,
	userService service.User) Query {
	return Query{
		taskService: taskService,
		executionService: executionService,
		userService: userService,
	}
}


