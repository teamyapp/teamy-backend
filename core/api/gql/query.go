package gql

import (
	"context"

	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Query struct {
	deps *Dependencies
}

func (q Query) Me(ct context.Context) (User, error) {
	user, err := q.deps.userService.Me(ct)
	if err != nil {
		q.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return User{}, err
	}

	return newUser(q.deps, user), nil
}

func (q Query) Tasks(ct context.Context, args struct {
	Filter *TaskFilter
}) ([]Task, error) {
	filter, err := fromGraphQLTaskFilterPtr(ct, q.deps.dataCollector, args.Filter)
	if err != nil {
		q.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	tasks, err := q.deps.taskService.FindTasks(ct, filter)
	if err != nil {
		q.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	return collect.Map(tasks, func(task entity.Task, _ int) Task {
		return newTask(q.deps, task)
	}), nil
}

func (q Query) Teams(ct context.Context, args struct {
	Filter *TeamFilter
}) ([]Team, error) {
	filter, err := fromGraphQLTeamFilterPtr(ct, q.deps.dataCollector, args.Filter)
	if err != nil {
		q.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	teams, err := q.deps.teamService.FindTeams(ct, filter)
	if err != nil {
		q.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	return collect.Map(teams, func(team entity.Team, _ int) Team {
		return newTeam(q.deps, team)
	}), nil
}

func (q Query) Invitations(ct context.Context, args struct {
	Filter *InvitationFilter
}) ([]Invitation, error) {
	filter, err := fromGraphQLInvitationFilterPtr(ct, q.deps.dataCollector, args.Filter)
	if err != nil {
		q.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	invitations, err := q.deps.invitationService.FindInvitations(ct, filter)
	if err != nil {
		q.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	return collect.Map(invitations, func(invitation entity.Invitation, _ int) Invitation {
		return newInvitation(q.deps, invitation)
	}), nil
}

func (q Query) Sprints(ct context.Context, args struct {
	Filter *SprintFilter
}) ([]Sprint, error) {
	filter, err := fromGraphQLSprintFilterPtr(ct, q.deps.dataCollector, args.Filter)
	if err != nil {
		q.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	sprints, err := q.deps.sprintService.FindSprints(ct, filter)
	if err != nil {
		q.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	return collect.Map(sprints, func(sprint entity.Sprint, _ int) Sprint {
		return newSprint(q.deps, sprint)
	}), nil
}

func (q Query) Apps(ct context.Context, args struct {
	Filter *AppFilter
}) ([]App, error) {
	panic("implement me")
}

func NewQuery(deps *Dependencies) Query {
	return Query{
		deps: deps,
	}
}
