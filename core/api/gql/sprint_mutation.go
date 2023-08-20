package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/service"
)

func (m Mutation) CreateSprint(ct context.Context, args struct {
	TeamID graphql.ID
	Sprint struct {
		StartAt graphql.Time
		EndAt   graphql.Time
	}
}) (Sprint, error) {
	teamID, argErr := fromGraphQLID(args.TeamID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Sprint{}, errs.ToResolverErr(internalErr)
	}

	input := service.CreateSprintInput{
		StartAt: args.Sprint.StartAt.Time,
		EndAt:   args.Sprint.EndAt.Time,
	}
	sprint, err := m.deps.sprintService.CreateSprint(ct, teamID, input)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Sprint{}, errs.ToResolverErr(err)
	}

	return newSprint(m.deps, sprint), nil
}

func (m Mutation) DeleteSprint(ct context.Context, args struct {
	SprintID graphql.ID
}) (Sprint, error) {
	sprintID, argErr := fromGraphQLID(args.SprintID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Sprint{}, errs.ToResolverErr(internalErr)
	}

	sprint, err := m.deps.sprintService.DeleteSprint(ct, sprintID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Sprint{}, errs.ToResolverErr(err)
	}

	return newSprint(m.deps, sprint), nil
}

func (m Mutation) AddTaskToSprint(ct context.Context, args struct {
	SprintID graphql.ID
	TaskID   graphql.ID
}) (Task, error) {
	sprintID, argErr := fromGraphQLID(args.SprintID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
	}

	taskID, argErr := fromGraphQLID(args.TaskID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
	}

	task, err := m.deps.sprintService.AddTaskToSprint(ct, sprintID, taskID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Task{}, errs.ToResolverErr(err)
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) RemoveTaskFromSprint(ct context.Context, args struct {
	SprintID graphql.ID
	TaskID   graphql.ID
}) (Task, error) {
	sprintID, argErr := fromGraphQLID(args.SprintID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
	}

	taskID, argErr := fromGraphQLID(args.TaskID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
	}

	task, err := m.deps.sprintService.RemoveTaskFromSprint(ct, sprintID, taskID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return Task{}, errs.ToResolverErr(err)
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) CopyTasksToSprint(ct context.Context, args struct {
	ToSprintID graphql.ID
	TaskIDs    []graphql.ID
}) ([]Task, error) {
	toSprintID, argErr := fromGraphQLID(args.ToSprintID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	taskIDs := make([]uint64, 0)
	for _, TaskID := range args.TaskIDs {
		taskID, err := fromGraphQLID(TaskID)
		if err != nil {
			internalErr := errs.NewError(
				errs.InvalidArgument,
				argErr.Error(),
			)
			m.deps.logger.ErrorWithContext(ct, internalErr)
			continue
		}

		taskIDs = append(taskIDs, taskID)
	}

	tasks, err := m.deps.sprintService.CopyTasksToSprint(ct, toSprintID, taskIDs)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return []Task{}, errs.ToResolverErr(err)
	}

	return collect.Map(tasks, func(task entity.Task, _ int) Task {
		return newTask(m.deps, task)
	}), nil
}

func (m Mutation) MoveTasksToSprint(ct context.Context, args struct {
	FromSprintID graphql.ID
	ToSprintID   graphql.ID
	TaskIDs      []graphql.ID
}) ([]Task, error) {
	fromSprintID, argErr := fromGraphQLID(args.FromSprintID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	toSprintID, argErr := fromGraphQLID(args.ToSprintID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	taskIDs := make([]uint64, 0)
	for _, TaskID := range args.TaskIDs {
		taskID, err := fromGraphQLID(TaskID)
		if err != nil {
			internalErr := errs.NewError(
				errs.InvalidArgument,
				argErr.Error(),
			)
			m.deps.logger.ErrorWithContext(ct, internalErr)
			continue
		}

		taskIDs = append(taskIDs, taskID)
	}

	tasks, err := m.deps.sprintService.MoveTasksToSprint(ct, fromSprintID, toSprintID, taskIDs)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return []Task{}, errs.ToResolverErr(err)
	}

	gqlTasks := make([]Task, 0)
	for _, task := range tasks {
		gqlTasks = append(gqlTasks, newTask(m.deps, task))
	}

	return gqlTasks, nil
}
