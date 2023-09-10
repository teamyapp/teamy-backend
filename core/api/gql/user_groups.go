package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/service"
)

type StaticUserGroup struct {
	deps  *Dependencies
	group entity.StaticGroup
}

var _ Group = (*StaticUserGroup)(nil)

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

	createUserStaticGroupInput := service.CreateUserStaticGroupInput{
		Name:    args.Input.Name,
		UserIDs: userIDs,
	}
	group, err := m.deps.groupService.CreateStaticUserGroup(ctx, appID, createUserStaticGroupInput)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return StaticUserGroup{}
	}

	return newStaticUserGroup(m.deps, group)
}

func (m Mutation) UpdateStaticUserGroup(
	ctx context.Context,
	args struct {
		GroupID graphql.ID
		Input   struct {
			Name    string
			UserIDs []graphql.ID
		}
	},
) (StaticUserGroup, error) {
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

	updateUserStaticInput := service.UpdateUserStaticGroupInput{
		Name:    args.Input.Name,
		UserIDs: userIDs,
	}

	group, err := m.deps.groupService.UpdateStaticUserGroup(ctx, groupID, updateUserStaticInput)
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

	createFilterUserGroupInput := service.CreateFilterGroupInput{
		Name:   args.Input.Name,
		Filter: args.Input.Filter,
	}

	filterGroup, err := m.deps.groupService.CreateFilterUserGroup(ctx, appID, createFilterUserGroupInput)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return FilterGroup{}, errs.ToResolverErr(err)
	}

	return newFilterGroup(m.deps, filterGroup), nil
}

func newStaticUserGroup(deps *Dependencies, group entity.StaticGroup) StaticUserGroup {
	return StaticUserGroup{
		deps:  deps,
		group: group,
	}
}
