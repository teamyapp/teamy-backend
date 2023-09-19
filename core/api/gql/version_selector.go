package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type VersionSelector interface {
	ID(ct context.Context) graphql.ID
	Type(ct context.Context) entity.VersionSelectorType
}

type ExperimentVersionSelector struct {
	versionSelector entity.VersionSelector
	deps            *Dependencies
}

var _ VersionSelector = (*ExperimentVersionSelector)(nil)

func (v ExperimentVersionSelector) ID(ct context.Context) graphql.ID {
	return toGraphQLID(v.versionSelector.ID)
}

func (v ExperimentVersionSelector) Type(ct context.Context) entity.VersionSelectorType {
	return v.versionSelector.Type
}

func (v ExperimentVersionSelector) VersionNumbers(ct context.Context) ([]int32, error) {
	versionNumbers, err := v.deps.rolloutService.FindVersionNumbersByExperimentVersionSelectorID(ct, v.versionSelector.ID)
	if err != nil {
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(versionNumbers, func(versionNumber int, index int) int32 {
		return int32(versionNumber)
	}), nil
}

func newExperimentVersionSelector(versionSelector entity.VersionSelector, deps *Dependencies) ExperimentVersionSelector {
	return ExperimentVersionSelector{versionSelector: versionSelector, deps: deps}
}

type StaticVersionSelector struct {
	versionSelector entity.VersionSelector
	deps            *Dependencies
}

var _ VersionSelector = (*StaticVersionSelector)(nil)

func (v StaticVersionSelector) ID(ct context.Context) graphql.ID {
	return toGraphQLID(v.versionSelector.ID)
}

func (v StaticVersionSelector) Type(ct context.Context) entity.VersionSelectorType {
	return v.versionSelector.Type
}

func (v StaticVersionSelector) VersionNumber(ct context.Context) (int32, error) {
	versionNumber, err := v.deps.rolloutService.FindVersionNumberByStaticVersionSelectorID(ct, v.versionSelector.ID)
	if err != nil {
		return 0, errs.ToResolverErr(err)
	}

	return int32(versionNumber), nil
}

func newStaticVersionSelector(versionSelector entity.VersionSelector, deps *Dependencies) StaticVersionSelector {
	return StaticVersionSelector{versionSelector: versionSelector, deps: deps}
}

func (m Mutation) CreateStaticVersionSelector(
	ct context.Context,
	args struct {
		Input struct {
			VersionNumber int32
		}
	},
) (StaticVersionSelector, error) {
	versionNumber := int(args.Input.VersionNumber)
	versionSelector, err := m.deps.rolloutService.CreateStaticVersionSelector(ct, versionNumber)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return StaticVersionSelector{}, errs.ToResolverErr(err)
	}

	return newStaticVersionSelector(versionSelector, m.deps), nil
}

func (m Mutation) CreateExperimentVersionSelector(
	ct context.Context,
	args struct {
		Input struct {
			VersionNumbers []int32
		}
	},
) (ExperimentVersionSelector, error) {
	versionNumbers := collect.Map(args.Input.VersionNumbers, func(versionNumber int32, index int) int {
		return int(versionNumber)
	})
	versionSelector, err := m.deps.rolloutService.CreateExperimentVersionSelector(ct, versionNumbers)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return ExperimentVersionSelector{}, errs.ToResolverErr(err)
	}

	return newExperimentVersionSelector(versionSelector, m.deps), nil
}
