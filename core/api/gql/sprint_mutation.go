package gql

import (
	"context"
	"log"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/service"
)

func (m Mutation) CreateSprint(ct context.Context, args struct {
	TeamID graphql.ID
	Sprint struct {
		StartAt graphql.Time
	}
}) (Sprint, error) {
	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		log.Println(err)
		return Sprint{}, err
	}

	input := service.CreateSprintInput{
		StartAt: args.Sprint.StartAt.Time,
	}

	sprint, err := m.deps.sprintService.CreateSprint(ct, teamID, input)
	if err != nil {
		log.Println(err)
		return Sprint{}, err
	}

	return newSprint(m.deps, sprint), nil
}

func (m Mutation) DeleteSprint(ct context.Context, args struct {
	SprintID graphql.ID
}) (Sprint, error) {
	sprintID, err := fromGraphQLID(args.SprintID)
	if err != nil {
		log.Println(err)
		return Sprint{}, err
	}

	sprint, err := m.deps.sprintService.DeleteSprint(ct, sprintID)
	if err != nil {
		log.Println(err)
		return Sprint{}, err
	}

	return newSprint(m.deps, sprint), nil
}

func (m Mutation) AddTaskToSprint(ct context.Context, args struct {
	SprintID graphql.ID
	TaskID   graphql.ID
}) (Task, error) {
	sprintID, err := fromGraphQLID(args.SprintID)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	task, err := m.deps.sprintService.AddTaskToSprint(ct, sprintID, taskID)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) RemoveTaskFromSprint(ct context.Context, args struct {
	SprintID graphql.ID
	TaskID   graphql.ID
}) (Task, error) {
	sprintID, err := fromGraphQLID(args.SprintID)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	task, err := m.deps.sprintService.RemoveTaskFromSprint(ct, sprintID, taskID)
	if err != nil {
		log.Println(err)
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}
