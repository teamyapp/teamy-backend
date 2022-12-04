package gql

import (
	"context"

	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/feature"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/service"
)

func (m Mutation) CreateSprint(ct context.Context, args struct {
	TeamID graphql.ID
	Sprint struct {
		StartAt graphql.Time
		EndAt   graphql.Time
	}
}) (Sprint, error) {
	teamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Sprint{}, err
	}

	input := service.CreateSprintInput{
		StartAt: args.Sprint.StartAt.Time,
		EndAt:   args.Sprint.EndAt.Time,
	}
	sprint, err := m.deps.sprintService.CreateSprint(ct, teamID, input)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Sprint{}, err
	}

	if feature.EnableAuthorization {
		err = m.registerResource(ct, authorization.SprintResourceType, sprint.ID)
		if err != nil {
			m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return Sprint{}, err
		}

		err = m.assignParentResource(ct, authorization.SprintResourceType, sprint.ID, authorization.TeamResourceType, sprint.OwningTeamID)
		if err != nil {
			m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return Sprint{}, err
		}
	}

	return newSprint(m.deps, sprint), nil
}

func (m Mutation) DeleteSprint(ct context.Context, args struct {
	SprintID graphql.ID
}) (Sprint, error) {
	sprintID, err := fromGraphQLID(args.SprintID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Sprint{}, err
	}

	sprint, err := m.deps.sprintService.DeleteSprint(ct, sprintID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.sprintService.AddTaskToSprint(ct, sprintID, taskID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.sprintService.RemoveTaskFromSprint(ct, sprintID, taskID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) CopyTasksToSprint(ct context.Context, args struct {
	ToSprintID graphql.ID
	TaskIDs    []graphql.ID
}) ([]Task, error) {
	toSprintID, err := fromGraphQLID(args.ToSprintID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return []Task{}, err
	}

	taskIDs := make([]uint64, 0)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return []Task{}, err
	}

	for _, TaskID := range args.TaskIDs {
		taskID, err := fromGraphQLID(TaskID)
		if err != nil {
			m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		taskIDs = append(taskIDs, taskID)
	}

	tasks, err := m.deps.sprintService.CopyTasksToSprint(ct, toSprintID, taskIDs)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return []Task{}, err
	}

	gqlTasks := make([]Task, 0)
	for _, task := range tasks {
		gqlTasks = append(gqlTasks, newTask(m.deps, task))
	}

	return gqlTasks, nil
}

func (m Mutation) MoveTasksToSprint(ct context.Context, args struct {
	FromSprintID graphql.ID
	ToSprintID   graphql.ID
	TaskIDs      []graphql.ID
}) ([]Task, error) {
	fromSprintID, err := fromGraphQLID(args.FromSprintID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return []Task{}, err
	}

	toSprintID, err := fromGraphQLID(args.ToSprintID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return []Task{}, err
	}

	taskIDs := make([]uint64, 0)
	for _, TaskID := range args.TaskIDs {
		taskID, err := fromGraphQLID(TaskID)
		if err != nil {
			m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		taskIDs = append(taskIDs, taskID)
	}

	tasks, err := m.deps.sprintService.MoveTasksToSprint(ct, fromSprintID, toSprintID, taskIDs)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return []Task{}, err
	}

	gqlTasks := make([]Task, 0)
	for _, task := range tasks {
		gqlTasks = append(gqlTasks, newTask(m.deps, task))
	}

	return gqlTasks, nil
}
