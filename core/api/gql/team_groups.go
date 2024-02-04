package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/service"
)

type StaticTeamGroup struct {
	deps  *Dependencies
	group entity.StaticGroup
}

var _ Group = (*StaticTeamGroup)(nil)

func (s StaticTeamGroup) ID(ctx context.Context) graphql.ID {
	return toGraphQLID(s.group.ID)
}

func (s StaticTeamGroup) Type(ctx context.Context) entity.GroupType {
	return s.group.Type
}

func (s StaticTeamGroup) MemberType(ctx context.Context) entity.GroupMemberType {
	return s.group.MemberType
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

func (s StaticTeamGroup) GroupRolloutRelations(ctx context.Context) ([]GroupRolloutRelation, error) {
	relations, err := s.deps.groupService.FindGroupRolloutRelationsByGroupID(ctx, s.group.ID)
	if err != nil {
		s.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(relations, func(relation entity.GroupRolloutRelation, index int) GroupRolloutRelation {
		return newGroupRolloutRelation(s.deps, relation)
	}), nil
}

func (s StaticTeamGroup) Apps(ctx context.Context) ([]App, error) {
	apps, err := s.deps.appService.FindAppsByGroupID(ctx, s.group.ID)
	if err != nil {
		s.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(apps, func(app entity.App, index int) App {
		return newApp(s.deps, app)
	}), nil
}

func (s StaticTeamGroup) ToStaticUserGroup() (*StaticUserGroup, bool) {
	return nil, false
}

func (s StaticTeamGroup) ToStaticTeamGroup() (*StaticTeamGroup, bool) {
	return &s, true
}

func (s StaticTeamGroup) ToFilterGroup() (*FilterGroup, bool) {
	return nil, false
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
	for _, id := range args.Input.TeamIDs {
		teamID, err := fromGraphQLID(id)
		if err != nil {
			internalErr := errs.NewError(
				errs.InvalidArgument,
				err.Error(),
			)
			m.deps.logger.ErrorWithContext(ctx, internalErr)
			return StaticTeamGroup{}, errs.ToResolverErr(internalErr)
		}

		teamIDs = append(teamIDs, teamID)
	}

	createStaticTeamGroupInput := service.CreateStaticGroupInput{
		Name:            args.Input.Name,
		MemberIDs:       teamIDs,
		GroupMemberType: entity.GroupMemberTypeTeam,
	}
	group, err := m.deps.groupService.CreateAppStaticGroup(ctx, appID, createStaticTeamGroupInput)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return StaticTeamGroup{}, errs.ToResolverErr(err)
	}

	return newStaticTeamGroup(m.deps, group), nil
}

func (m Mutation) UpdateStaticTeamGroup(
	ctx context.Context,
	args struct {
		GroupID graphql.ID
		Input   struct {
			Name    string
			TeamIDs []graphql.ID
		}
	},
) (StaticTeamGroup, error) {
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
	for _, id := range args.Input.TeamIDs {
		teamID, err := fromGraphQLID(id)
		if err != nil {
			internalErr := errs.NewError(
				errs.InvalidArgument,
				err.Error(),
			)
			m.deps.logger.ErrorWithContext(ctx, internalErr)
			return StaticTeamGroup{}, errs.ToResolverErr(internalErr)
		}

		teamIDs = append(teamIDs, teamID)
	}

	updateStaticTeamGroupInput := service.UpdateStaticGroupInput{
		Name:            args.Input.Name,
		MemberIDs:       teamIDs,
		GroupMemberType: entity.GroupMemberTypeTeam,
	}
	group, err := m.deps.groupService.UpdateStaticGroup(ctx, groupID, updateStaticTeamGroupInput)
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
	},
) (FilterGroup, error) {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return FilterGroup{}, errs.ToResolverErr(internalErr)
	}

	createFilterGroupInput := service.CreateFilterGroupInput{
		Name:            args.Input.Name,
		Filter:          args.Input.Filter,
		GroupMemberType: entity.GroupMemberTypeTeam,
	}

	group, err := m.deps.groupService.CreateFilterGroup(ctx, appID, createFilterGroupInput)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return FilterGroup{}, errs.ToResolverErr(err)
	}

	return newFilterGroup(m.deps, group), nil
}

func newStaticTeamGroup(deps *Dependencies, group entity.StaticGroup) StaticTeamGroup {
	return StaticTeamGroup{
		deps:  deps,
		group: group,
	}
}
