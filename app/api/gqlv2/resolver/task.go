package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

var availableActions = map[entityv2.TaskStatus][]entityv2.TaskAction{
	entityv2.TaskStatusUpcoming: {
		entityv2.TaskActionStart,
		entityv2.TaskActionDelete,
		entityv2.TaskActionAssignOwner,
	},
	entityv2.TaskStatusInProgress: {
		entityv2.TaskActionMarkComplete,
		entityv2.TaskActionReportBlocked,
		entityv2.TaskActionAssignOwner,
		entityv2.TaskActionDelete,
	},
	entityv2.TaskStatusDelivered: {
		entityv2.TaskActionDelete,
		entityv2.TaskActionAssignOwner,
	},
}

type Task struct {
	deps *Dependencies
	task entityv2.Task
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
	user, err := t.deps.userDao.FindUserByID(t.task.CreatorUserID)
	if err != nil {
		return User{}, err
	}

	return newUser(t.deps, user), nil
}

func (t Task) Owner(ct context.Context) (*User, error) {
	if t.task.OwnerUserID == nil {
		return nil, nil
	}

	owner, err := t.deps.userDao.FindUserByID(*t.task.OwnerUserID)
	if err != nil {
		return nil, err
	}

	gqlUser := newUser(t.deps, owner)
	return &gqlUser, nil
}

func (t Task) OwningTeam(ct context.Context) (*Team, error) {
	team, err := t.deps.teamDao.FindTeamByID(t.task.OwningTeamID)
	if err != nil {
		return nil, err
	}

	gqlTeam := newTeam(t.deps, team)
	return &gqlTeam, nil
}

func (t Task) Status(ct context.Context) entityv2.TaskStatus {
	return t.task.Status
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

func (t Task) DueAt(ct context.Context) *graphql.Time {
	return toGraphQLTimePtr(t.task.DueAt)
}

func (t Task) AvailableActions(ct context.Context) []entityv2.TaskAction {
	return availableActions[t.task.Status]
}

func newTask(deps *Dependencies, task entityv2.Task) Task {
	return Task{
		deps: deps,
		task: task,
	}
}
