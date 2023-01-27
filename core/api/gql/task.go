package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/api/gql/scalar"
	"github.com/teamyapp/teamy-backend/core/entity"
)

var availableActions = map[entity.TaskStatus][]entity.TaskAction{
	entity.TaskStatusTodo: {
		entity.TaskActionStart,
		entity.TaskActionDelete,
		entity.TaskActionAssignOwner,
	},
	entity.TaskStatusPaused: {
		entity.TaskActionStart,
		entity.TaskActionDelete,
		entity.TaskActionAssignOwner,
	},
	entity.TaskStatusInProgress: {
		entity.TaskActionMarkComplete,
		entity.TaskActionReportBlocked,
		entity.TaskActionAssignOwner,
		entity.TaskActionDelete,
	},
	entity.TaskStatusAwaiting: {
		entity.TaskActionAssignOwner,
		entity.TaskActionDelete,
	},
	entity.TaskStatusDelivered: {
		entity.TaskActionDelete,
		entity.TaskActionAssignOwner,
	},
}

type Task struct {
	deps *Dependencies
	task entity.Task
}

func (t Task) ID(ct context.Context) graphql.ID {
	return toGraphQLID(t.task.ID)
}

func (t Task) Goal(ct context.Context) string {
	return t.task.Goal
}

func (t Task) Context(ct context.Context) *string {
	return t.task.Context
}

func (t Task) Creator(ct context.Context) (User, error) {
	user, err := t.deps.userService.FindUserByID(ct, t.task.CreatorUserID)
	if err != nil {
		t.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return User{}, err
	}

	return newUser(t.deps, user), nil
}

func (t Task) Owner(ct context.Context) (*User, error) {
	if t.task.OwnerUserID == nil {
		return nil, nil
	}

	owner, err := t.deps.userService.FindUserByID(ct, *t.task.OwnerUserID)
	if err != nil {
		t.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	gqlUser := newUser(t.deps, owner)
	return &gqlUser, nil
}

func (t Task) OwningTeam(ct context.Context) (*Team, error) {
	team, err := t.deps.teamService.FindTeamByID(ct, t.task.OwningTeamID)
	if err != nil {
		t.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	gqlTeam := newTeam(t.deps, team)
	return &gqlTeam, nil
}

func (t Task) Status(ct context.Context) entity.TaskStatus {
	return t.task.Status
}

func (t Task) IsPlanned(ct context.Context) bool {
	return t.task.IsPlanned
}

func (t Task) Comments(ct context.Context) Thread {
	return newThread(t.deps, t.task.CommentsThreadID)
}

func (t Task) CreatedAt(ct context.Context) graphql.Time {
	return toGraphQLTime(t.task.CreatedAt)
}

func (t Task) UpdatedAt(ct context.Context) *graphql.Time {
	return toGraphQLTimePtr(t.task.UpdatedAt)
}

func (t Task) DeliveredAt(ct context.Context) *graphql.Time {
	return toGraphQLTimePtr(t.task.DeliveredAt)
}

func (t Task) DueAt(ct context.Context) *graphql.Time {
	return toGraphQLTimePtr(t.task.DueAt)
}

func (t Task) Effort(ct context.Context) *scalar.Duration {
	return toGraphQLDurationPtr(t.task.Effort)
}

func (t Task) AvailableActions(ct context.Context) []entity.TaskAction {
	return availableActions[t.task.Status]
}

func (t Task) AwaitForTasks(ct context.Context) ([]Task, error) {
	tasks, err := t.deps.taskService.FindAwaitForTasks(ct, t.task.ID)
	if err != nil {
		t.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	return collect.Map(tasks, func(task entity.Task, _ int) Task {
		return newTask(t.deps, task)
	}), nil
}

func (t Task) Links(ct context.Context) ([]TaskLink, error) {
	links, err := t.deps.taskLinkService.FindLinksByTaskID(ct, t.task.ID)
	if err != nil {
		t.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	return collect.Map(links, func(taskLink entity.TaskLink, _ int) TaskLink {
		return newTaskLink(t.deps, taskLink)
	}), nil
}

func newTask(deps *Dependencies, task entity.Task) Task {
	return Task{
		deps: deps,
		task: task,
	}
}
