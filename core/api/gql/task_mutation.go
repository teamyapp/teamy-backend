package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/api/gql/scalar"
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
	owningTeamID, argErr := fromGraphQLID(args.TeamID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
	}

	ownerUserID, argErr := fromGraphQLIDPtr(args.Task.OwnerUserID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
	}

	task, err := m.deps.taskService.CreateTask(ct, owningTeamID, service.CreateTaskInput{
		Goal:        args.Task.Goal,
		Context:     args.Task.Context,
		OwnerUserID: ownerUserID,
		DueAt:       fromGraphQLTimePtr(args.Task.DueAt),
		IsPlanned:   args.Task.IsPlanned,
	})
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return Task{}, errs.ToResolverErr(err)
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
	taskID, argErr := fromGraphQLID(args.TaskID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
	}

	ownerUserID, argErr := fromGraphQLIDPtr(args.Input.OwnerUserID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
	}

	owningTeamID, argErr := fromGraphQLID(args.Input.OwningTeamID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
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
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return Task{}, errs.ToResolverErr(err)
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) DeleteTask(ct context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	taskID, argErr := fromGraphQLID(args.TaskID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
	}

	task, err := m.deps.taskService.DeleteTask(ct, taskID)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return Task{}, errs.ToResolverErr(err)
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) MoveTaskToUpcoming(ct context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	taskID, argErr := fromGraphQLID(args.TaskID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
	}

	task, err := m.deps.taskService.MoveTaskToUpcoming(ct, taskID, true)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return Task{}, errs.ToResolverErr(err)
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) MoveTaskToInProgress(ct context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	taskID, argErr := fromGraphQLID(args.TaskID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
	}

	task, err := m.deps.taskService.MoveTaskToInProgress(ct, taskID)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return Task{}, errs.ToResolverErr(err)
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) MoveTaskToDelivered(ct context.Context, args struct {
	TaskID graphql.ID
}) (Task, error) {
	taskID, argErr := fromGraphQLID(args.TaskID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
	}

	task, err := m.deps.taskService.MoveTaskToDelivered(ct, taskID)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return Task{}, errs.ToResolverErr(err)
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) MoveTaskToBlocked(ct context.Context, args struct {
	TaskID graphql.ID
	Reason string
}) (Task, error) {
	taskID, argErr := fromGraphQLID(args.TaskID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
	}

	task, err := m.deps.taskService.MoveTaskToBlocked(ct, taskID, args.Reason)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return Task{}, errs.ToResolverErr(err)
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) AddAwaitForTask(ct context.Context, args struct {
	TaskID         graphql.ID
	AwaitForTaskId graphql.ID
}) (Task, error) {
	taskID, argErr := fromGraphQLID(args.TaskID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
	}

	awaitForTaskId, argErr := fromGraphQLID(args.AwaitForTaskId)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
	}

	task, err := m.deps.taskService.AddAwaitForTask(ct, taskID, awaitForTaskId)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return Task{}, errs.ToResolverErr(err)
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) RemoveAwaitForTask(ct context.Context, args struct {
	TaskID         graphql.ID
	AwaitForTaskId graphql.ID
}) (Task, error) {
	taskID, argErr := fromGraphQLID(args.TaskID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
	}

	awaitForTaskId, argErr := fromGraphQLID(args.AwaitForTaskId)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return Task{}, errs.ToResolverErr(internalErr)
	}

	task, err := m.deps.taskService.RemoveAwaitForTask(ct, taskID, awaitForTaskId)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return Task{}, errs.ToResolverErr(err)
	}

	return newTask(m.deps, task), nil
}

func (m Mutation) StartDraggingTask(ct context.Context, args struct {
	TaskID   graphql.ID
	ClientID graphql.ID
}) (graphql.ID, error) {
	taskID, argErr := fromGraphQLID(args.TaskID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", errs.ToResolverErr(internalErr)
	}

	clientID, argErr := fromGraphQLID(args.ClientID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", errs.ToResolverErr(internalErr)
	}

	err := m.deps.taskService.StartDraggingTask(ct, taskID, clientID)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return "", errs.ToResolverErr(err)
	}

	return args.TaskID, nil
}

func (m Mutation) StopDraggingTask(ct context.Context, args struct {
	TaskID   graphql.ID
	ClientID graphql.ID
}) (graphql.ID, error) {
	taskID, argErr := fromGraphQLID(args.TaskID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", errs.ToResolverErr(internalErr)
	}

	clientID, argErr := fromGraphQLID(args.ClientID)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		m.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return "", errs.ToResolverErr(internalErr)
	}

	err := m.deps.taskService.StopDraggingTask(ct, taskID, clientID)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return "", errs.ToResolverErr(err)
	}

	return args.TaskID, nil
}
