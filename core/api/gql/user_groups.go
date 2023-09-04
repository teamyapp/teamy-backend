package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type UserGroup interface {
	Group
	Rollouts(ctx context.Context) ([]Rollout, error)
}

type StaticUserGroup struct {
	deps  *Dependencies
	group entity.Group
}

var _ UserGroup = (*StaticUserGroup)(nil)

func (s StaticUserGroup) ID(ctx context.Context) graphql.ID {
	return toGraphQLID(s.group.ID)
}

func (s StaticUserGroup) Type(ctx context.Context) entity.GroupType {
	return s.group.Type
}

func (s StaticUserGroup) Name(ctx context.Context) string {
	return s.group.Name
}

func (s StaticUserGroup) Users(ctx context.Context) ([]User, error) {
	users, err := s.deps.groupService.FindUsersByGroupID(ctx, s.group.ID)

	if err != nil {
		s.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(users, func(user entity.User, index int) User {
		return newUser(s.deps, user)
	}), nil
}

func (s StaticUserGroup) CreatedAt(ctx context.Context) graphql.Time {
	return toGraphQLTime(s.group.CreatedAt)
}

func (s StaticUserGroup) UpdatedAt(ctx context.Context) *graphql.Time {
	return toGraphQLTimePtr(s.group.UpdatedAt)
}

func (s StaticUserGroup) Rollouts(ctx context.Context) ([]Rollout, error) {
	rollouts, err := s.deps.rolloutService.FindRolloutsByGroupID(ctx, s.group.ID)
	if err != nil {
		s.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(rollouts, func(rollout entity.Rollout, index int) Rollout {
		return newRollout(s.deps, rollout)
	}), nil
}

func (s StaticUserGroup) App(ctx context.Context) (App, error) {
	app, err := s.deps.appService.FindAppByID(ctx, s.group.ID)
	if err != nil {
		s.deps.logger.ErrorWithContext(ctx, err)
		return App{}, errs.ToResolverErr(err)
	}

	return newApp(s.deps, app), nil
}

type FilterUserGroup struct {
	deps        *Dependencies
	filterGroup entity.FilterGroup
}

var _ UserGroup = (*FilterUserGroup)(nil)
var _ FilterGroup = (*FilterUserGroup)(nil)

func (f FilterUserGroup) ID(ctx context.Context) graphql.ID {
	return toGraphQLID(f.filterGroup.ID)
}

func (f FilterUserGroup) Type(ctx context.Context) entity.GroupType {
	return f.filterGroup.Type
}

func (f FilterUserGroup) Name(ctx context.Context) string {
	return f.filterGroup.Name
}

func (f FilterUserGroup) Filter(ctx context.Context) string {
	return f.filterGroup.Filter
}

func (f FilterUserGroup) UserCount(ctx context.Context) int32 {
	return int32(f.filterGroup.Count)
}

func (f FilterUserGroup) CreatedAt(ctx context.Context) graphql.Time {
	return toGraphQLTime(f.filterGroup.CreatedAt)
}

func (f FilterUserGroup) UpdatedAt(ctx context.Context) *graphql.Time {
	return toGraphQLTimePtr(f.filterGroup.UpdatedAt)
}

func (f FilterUserGroup) Rollouts(ctx context.Context) ([]Rollout, error) {
	rollouts, err := f.deps.rolloutService.FindRolloutsByGroupID(ctx, f.filterGroup.ID)
	if err != nil {
		f.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(rollouts, func(rollout entity.Rollout, index int) Rollout {
		return newRollout(f.deps, rollout)
	}), nil
}

func (f FilterUserGroup) App(ctx context.Context) App {
	app, err := f.deps.appService.FindAppByID(ctx, f.filterGroup.ID)
	if err != nil {
		f.deps.logger.ErrorWithContext(ctx, err)
		return App{}
	}

	return newApp(f.deps, app)
}

func (m Mutation) CreateStaticUserGroup(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			Name    string
			UserIDs []graphql.ID
		}
	},
) StaticUserGroup {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return StaticUserGroup{}
	}

	userIDs := make([]uint64, len(args.Input.UserIDs))
	for _, userID := range args.Input.UserIDs {
		id, err := fromGraphQLID(userID)
		if err != nil {
			internalErr := errs.NewError(
				errs.InvalidArgument,
				err.Error(),
			)
			m.deps.logger.ErrorWithContext(ctx, internalErr)
			return StaticUserGroup{}
		}
		userIDs = append(userIDs, id)
	}

	group, err := m.deps.groupService.CreateStaticUserGroup(ctx, appID, args.Input.Name, userIDs)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return StaticUserGroup{}
	}

	return newStaticUserGroup(m.deps, group)
}

func (m Mutation) UpdateStaticUserGroup(
	ctx context.Context,
	args struct {
		AppID   graphql.ID
		GroupID graphql.ID
		Input   struct {
			Name    string
			UserIDs []graphql.ID
		}
	},
) (StaticUserGroup, error) {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return StaticUserGroup{}, errs.ToResolverErr(internalErr)
	}

	groupID, internalErr := fromGraphQLID(args.GroupID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return StaticUserGroup{}, errs.ToResolverErr(internalErr)
	}

	userIDs := make([]uint64, len(args.Input.UserIDs))
	for _, userID := range args.Input.UserIDs {
		id, err := fromGraphQLID(userID)
		if err != nil {
			internalErr := errs.NewError(
				errs.InvalidArgument,
				err.Error(),
			)
			m.deps.logger.ErrorWithContext(ctx, internalErr)
			return StaticUserGroup{}, errs.ToResolverErr(internalErr)
		}
		userIDs = append(userIDs, id)
	}

	group, err := m.deps.groupService.UpdateStaticUserGroup(ctx, appID, groupID, args.Input.Name, userIDs)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return StaticUserGroup{}, errs.ToResolverErr(err)
	}

	return newStaticUserGroup(m.deps, group), nil
}

func (m Mutation) CreateFilterUserGroup(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			Name   string
			Filter string
		}
	},
) (FilterUserGroup, error) {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return FilterUserGroup{}, errs.ToResolverErr(internalErr)
	}

	filterGroup, err := m.deps.groupService.CreateFilterUserGroup(ctx, appID, args.Input.Name, args.Input.Filter)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return FilterUserGroup{}, errs.ToResolverErr(err)
	}

	return newFilterUserGroup(m.deps, filterGroup), nil
}

func (m Mutation) UpdateFilterUserGroup(
	ctx context.Context,
	args struct {
		AppID   graphql.ID
		GroupID graphql.ID
		Input   struct {
			Name   string
			Filter string
		}
	},
) (FilterUserGroup, error) {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return FilterUserGroup{}, errs.ToResolverErr(internalErr)
	}

	groupID, internalErr := fromGraphQLID(args.GroupID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return FilterUserGroup{}, errs.ToResolverErr(internalErr)
	}

	filterGroup, err := m.deps.groupService.UpdateFilterUserGroup(ctx, appID, groupID, args.Input.Name, args.Input.Filter)

	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return FilterUserGroup{}, errs.ToResolverErr(err)
	}

	return newFilterUserGroup(m.deps, filterGroup), nil
}

func newFilterUserGroup(deps *Dependencies, filterGroup entity.FilterGroup) FilterUserGroup {
	return FilterUserGroup{
		deps:        deps,
		filterGroup: filterGroup,
	}
}

func newStaticUserGroup(deps *Dependencies, group entity.Group) StaticUserGroup {
	return StaticUserGroup{
		deps:  deps,
		group: group,
	}
}
