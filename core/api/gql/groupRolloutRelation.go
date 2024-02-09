package gql

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type GroupRolloutRelation struct {
	deps                 *Dependencies
	groupRolloutRelation entity.GroupRolloutRelation
}

func (g GroupRolloutRelation) Group(ctx context.Context) (Group, error) {
	group, err := g.deps.groupService.FindGroupByID(ctx, g.groupRolloutRelation.GroupID)
	if err != nil {
		g.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	switch group.Type {
	case entity.GroupTypeStatic:
		switch group.MemberType {
		case entity.GroupMemberTypeUser:
			return newStaticUserGroup(g.deps, group.StaticGroup), nil
		case entity.GroupMemberTypeTeam:
			return newStaticTeamGroup(g.deps, group.StaticGroup), nil
		default:
			return nil, errs.ToResolverErr(errs.NewError(errs.Unknown, "unknown group member type"))
		}

	case entity.GroupTypeFilter:
		return newFilterGroup(g.deps, group.FilterGroup), nil
	default:
		return nil, errs.ToResolverErr(errs.NewError(errs.Unknown, "unknown group type"))
	}
}

func (g GroupRolloutRelation) Rollout(ctx context.Context) (Rollout, error) {
	rollout, err := g.deps.rolloutService.FindRolloutByID(ctx, g.groupRolloutRelation.RolloutID)
	if err != nil {
		g.deps.logger.ErrorWithContext(ctx, err)
		return Rollout{}, errs.ToResolverErr(err)
	}

	return newRollout(g.deps, rollout), nil
}

func (g GroupRolloutRelation) OrderIndex() int32 {
	return int32(g.groupRolloutRelation.OrderIndex)
}

func newGroupRolloutRelation(deps *Dependencies, groupRolloutRelation entity.GroupRolloutRelation) GroupRolloutRelation {
	return GroupRolloutRelation{
		deps:                 deps,
		groupRolloutRelation: groupRolloutRelation,
	}
}
