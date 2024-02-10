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
	MemberType(ctx context.Context) entity.GroupMemberType
	MaxRolloutIndex(ctx context.Context) int32
	Name(ctx context.Context) string
	CreatedAt(ctx context.Context) graphql.Time
	UpdatedAt(ctx context.Context) *graphql.Time
	GroupRolloutRelations(ctx context.Context) ([]GroupRolloutRelation, error)
	Apps(ctx context.Context) ([]App, error)
	ToStaticUserGroup() (*StaticUserGroup, bool)
	ToStaticTeamGroup() (*StaticTeamGroup, bool)
	ToFilterGroup() (*FilterGroup, bool)
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

func (f FilterGroup) MemberType(ctx context.Context) entity.GroupMemberType {
	return f.filterGroup.MemberType
}

func (f FilterGroup) MaxRolloutIndex(ctx context.Context) int32 {
	return int32(f.filterGroup.MaxRolloutIndex)
}

func (f FilterGroup) Name(ctx context.Context) string {
	return f.filterGroup.Name
}

func (f FilterGroup) Filter(ctx context.Context) string {
	return f.filterGroup.Filter
}

func (f FilterGroup) CreatedAt(ctx context.Context) graphql.Time {
	return toGraphQLTime(f.filterGroup.CreatedAt)
}

func (f FilterGroup) UpdatedAt(ctx context.Context) *graphql.Time {
	return toGraphQLTimePtr(f.filterGroup.UpdatedAt)
}

func (f FilterGroup) GroupRolloutRelations(ctx context.Context) ([]GroupRolloutRelation, error) {
	relations, err := f.deps.groupService.FindGroupRolloutRelationsByGroupID(ctx, f.filterGroup.ID)
	if err != nil {
		f.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(relations, func(relation entity.GroupRolloutRelation, index int) GroupRolloutRelation {
		return newGroupRolloutRelation(f.deps, relation)
	}), nil
}

func (f FilterGroup) Apps(ctx context.Context) ([]App, error) {
	apps, err := f.deps.appService.FindAppsByGroupID(ctx, f.filterGroup.ID)
	if err != nil {
		f.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(apps, func(app entity.App, index int) App {
		return newApp(f.deps, app)
	}), nil
}

func (f FilterGroup) ToStaticUserGroup() (*StaticUserGroup, bool) {
	return nil, false
}

func (f FilterGroup) ToStaticTeamGroup() (*StaticTeamGroup, bool) {
	return nil, false
}

func (f FilterGroup) ToFilterGroup() (*FilterGroup, bool) {
	return &f, true
}

func newFilterGroup(deps *Dependencies, filterGroup entity.FilterGroup) FilterGroup {
	return FilterGroup{
		deps:        deps,
		filterGroup: filterGroup,
	}
}

func (m Mutation) CreateFilterGroup(
	ctx context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			Name            string
			Filter          string
			GroupMemberType entity.GroupMemberType
			RolloutIDs      []graphql.ID
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

	rolloutIDs := make([]uint64, len(args.Input.RolloutIDs))
	for index, id := range args.Input.RolloutIDs {
		rolloutID, err := fromGraphQLID(id)
		if err != nil {
			internalErr := errs.NewError(
				errs.InvalidArgument,
				err.Error(),
			)
			m.deps.logger.ErrorWithContext(ctx, internalErr)
			return FilterGroup{}, errs.ToResolverErr(internalErr)
		}

		rolloutIDs[index] = rolloutID
	}

	createFilterGroupInput := service.CreateFilterGroupInput{
		Name:            args.Input.Name,
		Filter:          args.Input.Filter,
		GroupMemberType: args.Input.GroupMemberType,
		RolloutIDs:      rolloutIDs,
	}

	filterGroup, err := m.deps.groupService.CreateFilterGroup(ctx, appID, createFilterGroupInput)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return FilterGroup{}, errs.ToResolverErr(err)
	}

	return newFilterGroup(m.deps, filterGroup), nil
}

func (m Mutation) UpdateGroup(
	ctx context.Context,
	args struct {
		AppID   graphql.ID
		GroupID graphql.ID
		Input   struct {
			Name            string
			Type            entity.GroupType
			GroupMemberType entity.GroupMemberType
			Filter          *string
			RolloutIDs      []graphql.ID
			MemberIDs       []graphql.ID
		}
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

	rolloutIDs := make([]uint64, len(args.Input.RolloutIDs))
	for index, id := range args.Input.RolloutIDs {
		rolloutID, err := fromGraphQLID(id)
		if err != nil {
			internalErr := errs.NewError(
				errs.InvalidArgument,
				err.Error(),
			)
			m.deps.logger.ErrorWithContext(ctx, internalErr)
			return nil, errs.ToResolverErr(internalErr)
		}

		rolloutIDs[index] = rolloutID
	}

	memberIDs := make([]uint64, len(args.Input.MemberIDs))
	for index, id := range args.Input.MemberIDs {
		memberID, err := fromGraphQLID(id)
		if err != nil {
			internalErr := errs.NewError(
				errs.InvalidArgument,
				err.Error(),
			)
			m.deps.logger.ErrorWithContext(ctx, internalErr)
			return nil, errs.ToResolverErr(internalErr)
		}

		memberIDs[index] = memberID
	}

	updateGroupInput := service.UpdateGroupInput{
		Name:            args.Input.Name,
		Filter:          args.Input.Filter,
		RolloutIDs:      rolloutIDs,
		MemberIDs:       memberIDs,
		GroupMemberType: args.Input.GroupMemberType,
		Type:            args.Input.Type,
	}

	group, err := m.deps.groupService.UpdateGroup(ctx, appID, groupID, updateGroupInput)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return getGroupFromGroupUnion(m.deps, group)
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

	return getGroupFromGroupUnion(m.deps, group)
}
