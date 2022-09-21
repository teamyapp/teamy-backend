package gql

import (
	"context"
	"errors"

	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Query struct {
	deps *Dependencies
}

func (q Query) Me(ct context.Context) (User, error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		q.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return User{}, err
	}

	user, err := q.deps.userDao.FindUserByID(ct, userID)
	if err != nil {
		q.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return User{}, err
	}

	return newUser(q.deps, user), nil
}

func (q Query) Tasks(ct context.Context, args struct {
	Filter *TaskFilter
}) ([]Task, error) {
	filter, err := fromGraphQLTaskFilterPtr(q.deps.dataCollector, args.Filter)
	if err != nil {
		q.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	tasks, err := q.deps.taskService.FindTasks(ct, filter)
	if err != nil {
		q.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return collect.Map(tasks, func(task entity.Task, _ int) Task {
		return newTask(q.deps, task)
	}), nil
}

func (q Query) Teams(ct context.Context, args struct {
	Filter *TeamFilter
}) ([]Team, error) {
	teams, err := q.deps.teamDao.FindAllTeams(ct)
	if err != nil {
		q.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	if args.Filter != nil {
		teams = collect.Filter(teams, func(team entity.Team) bool {
			return matchTeam(q.deps.dataCollector, *args.Filter, team)
		})
	}

	return collect.Map(teams, func(team entity.Team, _ int) Team {
		return newTeam(q.deps, team)
	}), nil
}

func (q Query) Invitations(ct context.Context, args struct {
	Filter *InvitationFilter
}) ([]Invitation, error) {
	invitations, err := q.deps.invitationDao.FindAllInvitations(ct)
	if err != nil {
		q.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	if args.Filter != nil {
		invitations = collect.Filter(invitations, func(invitation entity.Invitation) bool {
			return matchInvitation(q.deps.dataCollector, *args.Filter, invitation)
		})
	}

	return collect.Map(invitations, func(invitation entity.Invitation, _ int) Invitation {
		return newInvitation(q.deps, invitation)
	}), nil
}

func (q Query) Sprints(ct context.Context, args struct {
	Filter *SprintFilter
}) ([]Sprint, error) {
	filter, err := fromGraphQLSprintFilterPtr(q.deps.dataCollector, args.Filter)
	if err != nil {
		q.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	sprints, err := q.deps.sprintService.FindSprints(ct, filter)
	if err != nil {
		q.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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
