package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type VersionSelector struct {
	versionSelector entity.VersionSelector
	deps            *Dependencies
}

func (v VersionSelector) ID(ct context.Context) graphql.ID {
	return toGraphQLID(v.versionSelector.ID)
}

func (v VersionSelector) Type(ct context.Context) entity.VersionSelectorType {
	return v.versionSelector.Type
}

func (v VersionSelector) VersionNumbers(ct context.Context) ([]int32, error) {
	versionNumbers, err := v.deps.rolloutService.FindVersionNumbersByVersionSelectorID(ct, v.versionSelector.ID)
	if err != nil {
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(versionNumbers, func(versionNumber int, index int) int32 {
		return int32(versionNumber)
	}), nil
}

func newVersionSelector(versionSelector entity.VersionSelector, deps *Dependencies) VersionSelector {
	return VersionSelector{versionSelector: versionSelector, deps: deps}
}

func (m Mutation) CreateStaticVersionSelector(
	ct context.Context,
	args struct {
		Input struct {
			VersionNumber int32
		}
	},
) (VersionSelector, error) {
	versionNumber := int(args.Input.VersionNumber)
	versionSelector, err := m.deps.rolloutService.CreateStaticVersionSelector(ct, versionNumber)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return VersionSelector{}, errs.ToResolverErr(err)
	}

	return newVersionSelector(versionSelector, m.deps), nil
}

func (m Mutation) CreateExperimentVersionSelector(
	ct context.Context,
	args struct {
		Input struct {
			VersionNumbers []int32
		}
	},
) (VersionSelector, error) {
	versionNumbers := collect.Map(args.Input.VersionNumbers, func(versionNumber int32, index int) int {
		return int(versionNumber)
	})
	versionSelector, err := m.deps.rolloutService.CreateExperimentVersionSelector(ct, versionNumbers)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return VersionSelector{}, errs.ToResolverErr(err)
	}

	return newVersionSelector(versionSelector, m.deps), nil
}
