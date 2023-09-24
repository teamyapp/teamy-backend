package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/service"
)

type Group interface {
	ID(ctx context.Context) graphql.ID
	Type(ctx context.Context) entity.GroupType
	Name(ctx context.Context) string
	CreatedAt(ctx context.Context) graphql.Time
	UpdatedAt(ctx context.Context) *graphql.Time
	Rollouts(ctx context.Context) ([]Rollout, error)
}

type FilterGroup struct {
	deps        *Dependencies
	filterGroup entity.FilterGroup
}

var _ Group = (*FilterGroup)(nil)

func (f FilterGroup) ID(ctx context.Context) graphql.ID {
	return toGraphQLID(f.filterGroup.ID)
}

func (f FilterGroup) Type(ctx context.Context) entity.GroupType {
	return f.filterGroup.Type
}

func (f FilterGroup) Name(ctx context.Context) string {
	return f.filterGroup.Name
}

func (f FilterGroup) Filter(ctx context.Context) string {
	return f.filterGroup.Filter
}

func (f FilterGroup) MemberCount(ctx context.Context) int32 {
	return int32(f.filterGroup.Count)
}

func (f FilterGroup) CreatedAt(ctx context.Context) graphql.Time {
	return toGraphQLTime(f.filterGroup.CreatedAt)
}

func (f FilterGroup) UpdatedAt(ctx context.Context) *graphql.Time {
	return toGraphQLTimePtr(f.filterGroup.UpdatedAt)
}

func (f FilterGroup) Rollouts(ctx context.Context) ([]Rollout, error) {
	rollouts, err := f.deps.rolloutService.FindRolloutsByGroupID(ctx, f.filterGroup.ID)
	if err != nil {
		f.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(rollouts, func(rollout entity.Rollout, index int) Rollout {
		return newRollout(f.deps, rollout)
	}), nil
}

func (f FilterGroup) App(ctx context.Context) (App, error) {
	app, err := f.deps.appService.FindAppByID(ctx, f.filterGroup.ID)
	if err != nil {
		f.deps.logger.ErrorWithContext(ctx, err)
		return App{}, errs.ToResolverErr(err)
	}

	return newApp(f.deps, app), nil
}

func newFilterGroup(deps *Dependencies, filterGroup entity.FilterGroup) FilterGroup {
	return FilterGroup{
		deps:        deps,
		filterGroup: filterGroup,
	}
}

func (m Mutation) UpdateFilterGroup(
	ctx context.Context,
	args struct {
		GroupID graphql.ID
		Input   struct {
			Name   string
			Filter string
		}
	},
) (FilterGroup, error) {
	groupID, internalErr := fromGraphQLID(args.GroupID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return FilterGroup{}, errs.ToResolverErr(internalErr)
	}

	updateFilterGroupInput := service.UpdateFilterGroupInput{
		Name:   args.Input.Name,
		Filter: args.Input.Filter,
	}

	filterGroup, err := m.deps.groupService.UpdateFilterGroup(ctx, groupID, updateFilterGroupInput)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return FilterGroup{}, errs.ToResolverErr(err)
	}

	return newFilterGroup(m.deps, filterGroup), nil
}

func (m Mutation) DeleteAppGroup(
	ctx context.Context,
	args struct {
		AppID   graphql.ID
		GroupID graphql.ID
	},
) (Group, error) {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	groupID, internalErr := fromGraphQLID(args.GroupID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	group, err := m.deps.groupService.DeleteAppGroup(ctx, appID, groupID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	switch group.Type {
	case entity.GroupTypeStatic:
		appGroupRelationType, err := m.deps.groupService.FindAppGroupRelationType(ctx, appID, groupID)
		if err != nil {
			m.deps.logger.ErrorWithContext(ctx, err)
			return nil, errs.ToResolverErr(err)
		}

		switch appGroupRelationType {
		case entity.AppGroupRelationTypeUser:
			return newStaticUserGroup(m.deps, group.StaticGroup), nil
		case entity.AppGroupRelationTypeTeam:
			return newStaticTeamGroup(m.deps, group.StaticGroup), nil
		default:
			return nil, errs.ToResolverErr(errs.NewError(errs.Unknown, "unknown app group relation type"))
		}
	case entity.GroupTypeFilter:
		return newFilterGroup(m.deps, group.FilterGroup), nil
	default:
		return nil, errs.ToResolverErr(errs.NewError(errs.Unknown, "unknown group type"))
	}
}
