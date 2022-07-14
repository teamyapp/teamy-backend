package gql

import (
	"context"
	"log"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Sprint struct {
	deps   *Dependencies
	sprint entity.Sprint
}

func (s Sprint) ID(ct context.Context) graphql.ID {
	return toGraphQLID(s.sprint.ID)
}

func (s Sprint) StartAt(ct context.Context) graphql.Time {
	return toGraphQLTime(s.sprint.StartAt)
}

func (s Sprint) EndAt(ct context.Context) graphql.Time {
	return toGraphQLTime(s.sprint.EndAt)
}

func (s Sprint) CreatedAt(ct context.Context) graphql.Time {
	return toGraphQLTime(s.sprint.CreatedAt)
}

func (s Sprint) Tasks(ct context.Context, args struct {
	Filter *TaskFilter
}) ([]Task, error) {
	filter, err := fromGraphQLTaskFilterPtr(args.Filter)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tasks, err := s.deps.sprintService.FindTasksInSprint(ct, s.sprint.ID, filter)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return collect.Map(tasks, func(task entity.Task, index int) Task {
		return newTask(s.deps, task)
	}), nil
}

func newSprint(deps *Dependencies, sprint entity.Sprint) Sprint {
	return Sprint{
		deps:   deps,
		sprint: sprint,
	}
}
