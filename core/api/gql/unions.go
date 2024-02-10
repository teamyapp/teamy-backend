package gql

import (
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

func getGroupFromGroupUnion(
	deps *Dependencies,
	group entity.GroupUnion,
) (Group, error) {
	switch group.Type {
	case entity.GroupTypeStatic:
		switch group.MemberType {
		case entity.GroupMemberTypeUser:
			return newStaticUserGroup(deps, group.StaticGroup), nil
		case entity.GroupMemberTypeTeam:
			return newStaticTeamGroup(deps, group.StaticGroup), nil
		default:
			return nil, errs.ToResolverErr(errs.NewError(errs.Unknown, "unknown group member type"))
		}

	case entity.GroupTypeFilter:
		return newFilterGroup(deps, group.FilterGroup), nil
	default:
		return nil, errs.ToResolverErr(errs.NewError(errs.Unknown, "unknown group type"))
	}
}

func getActivatorFromActivatorUnion(
	deps *Dependencies,
	activator entity.ActivatorUnion,
) (Activator, error) {
	switch activator.Type {
	case entity.ActivatorTypeStatic:
		return newStaticActivator(activator.StaticActivator), nil
	case entity.ActivatorTypeTimeRange:
		return newTimeRangeActivator(activator.TimeRangeActivator), nil
	case entity.ActivatorTypeMaxViewers:
		return newMaxViewersActivator(activator.MaxViewersActivator), nil
	case entity.ActivatorTypePercentage:
		return newPercentageActivator(activator.PercentageActivator), nil
	default:
		return nil, errs.ToResolverErr(errs.NewError(errs.Unknown, fmt.Sprintf("Unknown activator type: %s", activator.Type)))
	}
}

func getVersionSelectorFromVersionSelectorUnion(
	deps *Dependencies,
	versionSelector entity.VersionSelectorUnion,
) (VersionSelector, error) {
	switch versionSelector.Type {
	case entity.VersionSelectorTypeExperiment:
		return newExperimentVersionSelector(versionSelector.ExperimentVersionSelector, deps), nil
	case entity.VersionSelectorTypeStatic:
		return newStaticVersionSelector(versionSelector.StaticVersionSelector, deps), nil
	default:
		return nil, errs.ToResolverErr(errs.NewError(errs.Unknown, fmt.Sprintf("Unknown version selector type: %s", versionSelector.Type)))
	}
}
