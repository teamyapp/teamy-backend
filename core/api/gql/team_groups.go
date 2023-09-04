package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamGroup interface {
	Group
	Rollouts(ctx context.Context) ([]Rollout, error)
}

type StaticTeamGroup struct {
	deps  *Dependencies
	group entity.Group
}

func (s StaticTeamGroup) ID(ctx context.Context) graphql.ID {
	return toGraphQLID(s.group.ID)
}

func (s StaticTeamGroup) Type(ctx context.Context) entity.GroupType {
	return s.group.Type
}

func (s StaticTeamGroup) Name(ctx context.Context) string {
	return s.group.Name
}

func (s StaticTeamGroup) Teams(ctx context.Context) ([]Team, error) {
	teams, err := s.deps.groupService.FindTeamsByGroupID(ctx, s.group.ID)

	if err != nil {
		s.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(teams, func(team entity.Team, index int) Team {
		return newTeam(s.deps, team)
	}), nil
}

func (s StaticTeamGroup) CreatedAt(ctx context.Context) graphql.Time {
	return toGraphQLTime(s.group.CreatedAt)
}

func (s StaticTeamGroup) UpdatedAt(ctx context.Context) *graphql.Time {
	return toGraphQLTimePtr(s.group.UpdatedAt)
}

func (s StaticTeamGroup) Rollouts(ctx context.Context) ([]Rollout, error) {
	rollouts, err := s.deps.rolloutService.FindRolloutsByGroupID(ctx, s.group.ID)
	if err != nil {
		s.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(rollouts, func(rollout entity.Rollout, index int) Rollout {
		return newRollout(s.deps, rollout)
	}), nil
}

func (s StaticTeamGroup) App(ctx context.Context) (App, error) {
	app, err := s.deps.appService.FindAppByID(ctx, s.group.ID)
	if err != nil {
		s.deps.logger.ErrorWithContext(ctx, err)
		return App{}, errs.ToResolverErr(err)
	}

	return newApp(s.deps, app), nil
}

type FilterTeamGroup struct {
	deps        *Dependencies
	filterGroup entity.FilterGroup
}

var _ TeamGroup = (*FilterTeamGroup)(nil)
var _ FilterGroup = (*FilterTeamGroup)(nil)

func (f FilterTeamGroup) ID(ctx context.Context) graphql.ID {
	return toGraphQLID(f.filterGroup.ID)
}

func (f FilterTeamGroup) Type(ctx context.Context) entity.GroupType {
	return f.filterGroup.Type
}

func (f FilterTeamGroup) Name(ctx context.Context) string {
	return f.filterGroup.Name
}

func (f FilterTeamGroup) Filter(ctx context.Context) string {
	return f.filterGroup.Filter
}

func (f FilterTeamGroup) TeamCount(ctx context.Context) int32 {
	return int32(f.filterGroup.Count)
}

func (f FilterTeamGroup) CreatedAt(ctx context.Context) graphql.Time {
	return toGraphQLTime(f.filterGroup.CreatedAt)
}

func (f FilterTeamGroup) UpdatedAt(ctx context.Context) *graphql.Time {
	return toGraphQLTimePtr(f.filterGroup.UpdatedAt)
}

func (f FilterTeamGroup) Rollouts(ctx context.Context) ([]Rollout, error) {
	rollouts, err := f.deps.rolloutService.FindRolloutsByGroupID(ctx, f.filterGroup.ID)
	if err != nil {
		f.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(rollouts, func(rollout entity.Rollout, index int) Rollout {
		return newRollout(f.deps, rollout)
	}), nil
}

func (f FilterTeamGroup) App(ctx context.Context) (App, error) {
	app, err := f.deps.appService.FindAppByID(ctx, f.filterGroup.ID)
	if err != nil {
		f.deps.logger.ErrorWithContext(ctx, err)
		return App{}, errs.ToResolverErr(err)
	}

	return newApp(f.deps, app), nil
}

func (m Mutation) CreateStaticTeamGroup(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			Name    string
			TeamIDs []graphql.ID
		}
	},
) (StaticTeamGroup, error) {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return StaticTeamGroup{}, errs.ToResolverErr(internalErr)
	}

	teamIDs := make([]uint64, len(args.Input.TeamIDs))
	for i, id := range args.Input.TeamIDs {
		teamID, err := fromGraphQLID(id)
		if err != nil {
			internalErr := errs.NewError(
				errs.InvalidArgument,
				err.Error(),
			)
			m.deps.logger.ErrorWithContext(ctx, internalErr)
			return StaticTeamGroup{}, errs.ToResolverErr(internalErr)
		}
		teamIDs[i] = teamID
	}

	group, err := m.deps.groupService.CreateStaticTeamGroup(ctx, appID, args.Input.Name, teamIDs)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return StaticTeamGroup{}, errs.ToResolverErr(err)
	}

	return newStaticTeamGroup(m.deps, group), nil
}

func (m Mutation) UpdateStaticTeamGroup(
	ctx context.Context,
	args struct {
		AppID   graphql.ID
		GroupID graphql.ID
		Input   struct {
			Name    string
			TeamIDs []graphql.ID
		}
	}) (StaticTeamGroup, error) {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return StaticTeamGroup{}, errs.ToResolverErr(internalErr)
	}
	groupID, internalErr := fromGraphQLID(args.GroupID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return StaticTeamGroup{}, errs.ToResolverErr(internalErr)
	}

	teamIDs := make([]uint64, len(args.Input.TeamIDs))
	for i, id := range args.Input.TeamIDs {
		teamID, err := fromGraphQLID(id)
		if err != nil {
			internalErr := errs.NewError(
				errs.InvalidArgument,
				err.Error(),
			)
			m.deps.logger.ErrorWithContext(ctx, internalErr)
			return StaticTeamGroup{}, errs.ToResolverErr(internalErr)
		}
		teamIDs[i] = teamID
	}

	group, err := m.deps.groupService.UpdateStaticTeamGroup(ctx, appID, groupID, args.Input.Name, teamIDs)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return StaticTeamGroup{}, errs.ToResolverErr(err)
	}

	return newStaticTeamGroup(m.deps, group), nil
}

func (m Mutation) CreateFilterTeamGroup(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			Name   string
			Filter string
		}
	}) (FilterTeamGroup, error) {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return FilterTeamGroup{}, errs.ToResolverErr(internalErr)
	}

	group, err := m.deps.groupService.CreateFilterTeamGroup(ctx, appID, args.Input.Name, args.Input.Filter)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return FilterTeamGroup{}, errs.ToResolverErr(err)
	}

	return newFilterTeamGroup(m.deps, group), nil
}

func (m Mutation) UpdateFilterTeamGroup(
	ctx context.Context,
	args struct {
		AppID   graphql.ID
		GroupID graphql.ID
		Input   struct {
			Name   string
			Filter string
		}
	},
) (FilterTeamGroup, error) {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return FilterTeamGroup{}, errs.ToResolverErr(internalErr)
	}

	groupID, internalErr := fromGraphQLID(args.GroupID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return FilterTeamGroup{}, errs.ToResolverErr(internalErr)
	}

	group, err := m.deps.groupService.UpdateFilterTeamGroup(ctx, appID, groupID, args.Input.Name, args.Input.Filter)

	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return FilterTeamGroup{}, errs.ToResolverErr(err)
	}

	return newFilterTeamGroup(m.deps, group), nil
}

func newStaticTeamGroup(deps *Dependencies, group entity.Group) StaticTeamGroup {
	return StaticTeamGroup{
		deps:  deps,
		group: group,
	}
}

func newFilterTeamGroup(deps *Dependencies, filterGroup entity.FilterGroup) FilterTeamGroup {
	return FilterTeamGroup{
		deps:        deps,
		filterGroup: filterGroup,
	}
}
