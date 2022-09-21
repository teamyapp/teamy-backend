package gql

import (
	"context"
	"errors"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/api/gql/scalar"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/service"
)

func (m Mutation) CreateTask(ct context.Context, args struct {
	TeamID graphql.ID
	Task   struct {
		Goal        string
		Context     *string
		OwnerUserID *graphql.ID
		DueAt       *graphql.Time
		IsPlanned   *bool
	}
}) (Task, error) {
	owningTeamID, err := fromGraphQLID(args.TeamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	ownerUserID, err := fromGraphQLIDPtr(m.deps.dataCollector, args.Task.OwnerUserID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.CreateTask(ct, owningTeamID, service.CreateTaskInput{
		Goal:        args.Task.Goal,
		Context:     args.Task.Context,
		OwnerUserID: ownerUserID,
		DueAt:       fromGraphQLTimePtr(args.Task.DueAt),
		IsPlanned:   args.Task.IsPlanned,
	})
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) UpdateTask(ct context.Context, args struct {
	TaskID graphql.ID
	Input  struct {
		Goal         string
		Context      *string
		OwnerUserID  *graphql.ID
		OwningTeamID graphql.ID
		Effort       *scalar.Duration
		DueAt        *graphql.Time
	}
}) (Task, error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	if feature.EnableAuthorization {
		query := authorization.NewUpdateTaskQuery(userID, taskID)
		hasPermission, err := m.hasPermission(ct, query)
		if err != nil {
			m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return Task{}, err
		}

		if !hasPermission {
			return Task{}, ResolverError{
				Code:    unauthorizedErrorCode,
				Message: fmt.Sprintf("Unauthorize: %v", query),
			}
		}
	}

	ownerUserID, err := fromGraphQLIDPtr(m.deps.dataCollector, args.Input.OwnerUserID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	owningTeamID, err := fromGraphQLID(args.Input.OwningTeamID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.UpdateTask(ct, taskID, service.UpdateTaskInput{
		Goal:         args.Input.Goal,
		Context:      args.Input.Context,
		OwnerUserID:  ownerUserID,
		OwningTeamID: owningTeamID,
		Effort:       fromGraphQLDurationPtr(args.Input.Effort),
		DueAt:        fromGraphQLTimePtr(args.Input.DueAt),
	})
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) DeleteTask(ct context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.DeleteTask(ct, taskID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) MoveTaskToUpcoming(ct context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.MoveTaskToUpcoming(ct, taskID, true)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) MoveTaskToInProgress(ct context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.MoveTaskToInProgress(ct, taskID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) MoveTaskToDelivered(ct context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.MoveTaskToDelivered(ct, taskID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) MoveTaskToBlocked(ct context.Context, args struct {
	TaskID graphql.ID
	Reason string
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.MoveTaskToBlocked(ct, taskID, args.Reason)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) AddAwaitForTask(ct context.Context, args struct {
	TaskID         graphql.ID
	AwaitForTaskId graphql.ID
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	awaitForTaskId, err := fromGraphQLID(args.AwaitForTaskId)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.AddAwaitForTask(ct, taskID, awaitForTaskId)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) RemoveAwaitForTask(ct context.Context, args struct {
	TaskID         graphql.ID
	AwaitForTaskId graphql.ID
}) (Task, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	awaitForTaskId, err := fromGraphQLID(args.AwaitForTaskId)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	task, err := m.deps.taskService.RemoveAwaitForTask(ct, taskID, awaitForTaskId)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Task{}, err
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) StartDraggingTask(ct context.Context, args struct {
	TaskID   graphql.ID
	ClientID graphql.ID
}) (graphql.ID, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	clientID, err := fromGraphQLID(args.ClientID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	err = m.deps.taskService.StartDraggingTask(ct, taskID, clientID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	return args.TaskID, nil
}

func (m Mutation) StopDraggingTask(ct context.Context, args struct {
	TaskID   graphql.ID
	ClientID graphql.ID
}) (graphql.ID, error) {
	taskID, err := fromGraphQLID(args.TaskID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	clientID, err := fromGraphQLID(args.ClientID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	err = m.deps.taskService.StopDraggingTask(ct, taskID, clientID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return "", err
	}

	return args.TaskID, nil
}
