package gql

import (
	"context"

	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Query struct {
	deps *Dependencies
}

func (q Query) Me(ct context.Context) (User, error) {
	user, err := q.deps.userService.Me(ct)
	if err != nil {
		q.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return User{}, errs.ToResolverErr(err)
	}

	return newUser(q.deps, user), nil
}

func (q Query) Tasks(ct context.Context, args struct {
	Filter *TaskFilter
}) ([]Task, error) {
	filter, argErr := fromGraphQLTaskFilterPtr(args.Filter)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		q.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	tasks, err := q.deps.taskService.FindTasks(ct, filter)
	if err != nil {
		q.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(tasks, func(task entity.Task, _ int) Task {
		return newTask(q.deps, task)
	}), nil
}

func (q Query) Teams(ct context.Context, args struct {
	Filter *TeamFilter
}) ([]Team, error) {
	filter, argErr := fromGraphQLTeamFilterPtr(args.Filter)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		q.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	teams, err := q.deps.teamService.FindTeams(ct, filter)
	if err != nil {
		q.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(teams, func(team entity.Team, _ int) Team {
		return newTeam(q.deps, team)
	}), nil
}

func (q Query) Invitations(ct context.Context, args struct {
	Filter *InvitationFilter
}) ([]Invitation, error) {
	filter, argErr := fromGraphQLInvitationFilterPtr(args.Filter)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		q.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	invitations, err := q.deps.invitationService.FindInvitations(ct, filter)
	if err != nil {
		q.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(invitations, func(invitation entity.Invitation, _ int) Invitation {
		return newInvitation(q.deps, invitation)
	}), nil
}

func (q Query) Sprints(ct context.Context, args struct {
	Filter *SprintFilter
}) ([]Sprint, error) {
	filter, argErr := fromGraphQLSprintFilterPtr(args.Filter)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		q.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	sprints, err := q.deps.sprintService.FindSprints(ct, filter)
	if err != nil {
		q.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(sprints, func(sprint entity.Sprint, _ int) Sprint {
		return newSprint(q.deps, sprint)
	}), nil
}

func (q Query) Apps(ct context.Context, args struct {
	Filter *AppFilter
}) ([]App, error) {
	filter, argErr := fromGraphQLAppFilterPtr(args.Filter)
	if argErr != nil {
		internalErr := &errs.Error{
			Code:     errs.InvalidArgument,
			EmbedErr: argErr,
		}
		q.deps.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	apps, err := q.deps.appService.FindApps(ct, filter)
	if err != nil {
		q.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(apps, func(app entity.App, _ int) App {
		return newApp(q.deps, app)
	}), nil
}

func NewQuery(deps *Dependencies) Query {
	return Query{
		deps: deps,
	}
}
