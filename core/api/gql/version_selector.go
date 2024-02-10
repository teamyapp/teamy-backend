package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/service"
)

type VersionSelector interface {
	ID(ct context.Context) graphql.ID
	Type(ct context.Context) entity.VersionSelectorType
	CreatedAt(ct context.Context) graphql.Time
	UpdatedAt(ct context.Context) *graphql.Time
	ToStaticVersionSelector() (*StaticVersionSelector, bool)
	ToExperimentVersionSelector() (*ExperimentVersionSelector, bool)
}

type StaticVersionSelector struct {
	versionSelector entity.StaticVersionSelector
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
	return int32(v.versionSelector.VersionNumber), nil
}

func (v StaticVersionSelector) CreatedAt(ct context.Context) graphql.Time {
	return toGraphQLTime(v.versionSelector.CreatedAt)
}

func (v StaticVersionSelector) UpdatedAt(ct context.Context) *graphql.Time {
	return toGraphQLTimePtr(v.versionSelector.UpdatedAt)
}

func (v StaticVersionSelector) ToStaticVersionSelector() (*StaticVersionSelector, bool) {
	return &v, true
}

func (v StaticVersionSelector) ToExperimentVersionSelector() (*ExperimentVersionSelector, bool) {
	return nil, false
}

func newStaticVersionSelector(versionSelector entity.StaticVersionSelector, deps *Dependencies) StaticVersionSelector {
	return StaticVersionSelector{versionSelector: versionSelector, deps: deps}
}

type ExperimentVersionSelector struct {
	versionSelector entity.ExperimentVersionSelector
	deps            *Dependencies
}

var _ VersionSelector = (*ExperimentVersionSelector)(nil)

func (v ExperimentVersionSelector) ID(ct context.Context) graphql.ID {
	return toGraphQLID(v.versionSelector.ID)
}

func (v ExperimentVersionSelector) Type(ct context.Context) entity.VersionSelectorType {
	return v.versionSelector.Type
}

func (v ExperimentVersionSelector) CreatedAt(ct context.Context) graphql.Time {
	return toGraphQLTime(v.versionSelector.CreatedAt)
}

func (v ExperimentVersionSelector) UpdatedAt(ct context.Context) *graphql.Time {
	return toGraphQLTimePtr(v.versionSelector.UpdatedAt)
}

func (v ExperimentVersionSelector) VersionNumbers(ct context.Context) []int32 {
	return collect.Map(v.versionSelector.VersionNumbers, func(versionNumber int, index int) int32 {
		return int32(versionNumber)
	})
}

func (v ExperimentVersionSelector) ToStaticVersionSelector() (*StaticVersionSelector, bool) {
	return nil, false
}

func (v ExperimentVersionSelector) ToExperimentVersionSelector() (*ExperimentVersionSelector, bool) {
	return &v, true
}

func newExperimentVersionSelector(versionSelector entity.ExperimentVersionSelector, deps *Dependencies) ExperimentVersionSelector {
	return ExperimentVersionSelector{versionSelector: versionSelector, deps: deps}
}

func (m Mutation) CreateStaticVersionSelector(
	ct context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			VersionNumber int32
		}
	},
) (StaticVersionSelector, error) {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return StaticVersionSelector{}, errs.ToResolverErr(internalErr)
	}

	versionNumber := int(args.Input.VersionNumber)
	versionSelector, err := m.deps.rolloutService.CreateStaticVersionSelector(ct, appID, versionNumber)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return StaticVersionSelector{}, errs.ToResolverErr(err)
	}

	return newStaticVersionSelector(versionSelector, m.deps), nil
}

func (m Mutation) UpdateVersionSelector(
	ctx context.Context,
	args struct {
		AppID             graphql.ID
		VersionSelectorID graphql.ID
		Input             struct {
			VersionNumber  *int32
			VersionNumbers []int32
			Type           entity.VersionSelectorType
		}
	},
) (VersionSelector, error) {
	versionSelectorID, internalErr := fromGraphQLID(args.VersionSelectorID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	updateVersionSelectorInput := service.UpdateVersionSelectorInput{
		VersionNumber: args.Input.VersionNumber,
		VersionNumbers: collect.Map(args.Input.VersionNumbers, func(versionNumber int32, index int) int {
			return int(versionNumber)
		}),
		Type: args.Input.Type,
	}

	versionSelector, err := m.deps.rolloutService.UpdateVersionSelector(ctx, appID, versionSelectorID, updateVersionSelectorInput)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return getVersionSelectorFromVersionSelectorUnion(m.deps, versionSelector)
}

func (m Mutation) CreateExperimentVersionSelector(
	ct context.Context,
	args struct {
		AppID graphql.ID
		Input struct {
			VersionNumbers []int32
		}
	},
) (ExperimentVersionSelector, error) {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return ExperimentVersionSelector{}, errs.ToResolverErr(internalErr)
	}

	versionNumbers := collect.Map(args.Input.VersionNumbers, func(versionNumber int32, index int) int {
		return int(versionNumber)
	})
	versionSelector, err := m.deps.rolloutService.CreateExperimentVersionSelector(ct, appID, versionNumbers)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return ExperimentVersionSelector{}, errs.ToResolverErr(err)
	}

	return newExperimentVersionSelector(versionSelector, m.deps), nil
}

func (m Mutation) DeleteVersionSelector(
	ctx context.Context,
	args struct {
		VersionSelectorID graphql.ID
	},
) (VersionSelector, error) {
	VersionSelectorID, internalErr := fromGraphQLID(args.VersionSelectorID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	versionSelector, err := m.deps.rolloutService.DeleteVersionSelector(ctx, VersionSelectorID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	return getVersionSelectorFromVersionSelectorUnion(m.deps, versionSelector)
}
