package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Rollout struct {
	deps    *Dependencies
	rollout entity.Rollout
}

func (r Rollout) ID() graphql.ID {
	return toGraphQLID(r.rollout.ID)
}

func (r Rollout) SelectorID() graphql.ID {
	return toGraphQLID(r.rollout.SelectorID)
}

func (r Rollout) SelectorType() entity.SelectorType {
	return r.rollout.SelectorType
}

func (r Rollout) ActivatorType() entity.ActivatorType {
	return r.rollout.ActivatorType
}

func (r Rollout) IsEnabled() bool {
	return r.rollout.IsEnabled
}

func (r Rollout) CreatedAt() graphql.Time {
	return toGraphQLTime(r.rollout.CreatedAt)
}

func (r Rollout) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(r.rollout.UpdatedAt)
}

func (r Rollout) Activator(ctx context.Context) (Activator, error) {
	switch r.rollout.ActivatorType {
	case entity.ActivatorTypeTimeRange:
		activator, err := r.deps.rolloutService.FindTimeRangeActivatorByID(ctx, r.rollout.ActivatorID)
		if err != nil {
			r.deps.logger.ErrorWithContext(ctx, err)
			return nil, errs.ToResolverErr(err)
		}

		return newTimeRangeActivator(r.deps, activator), nil
	case entity.ActivatorTypeMaxViewers:
		activator, err := r.deps.rolloutService.FindMaxViewersActivatorByID(ctx, r.rollout.ActivatorID)
		if err != nil {
			r.deps.logger.ErrorWithContext(ctx, err)
			return nil, errs.ToResolverErr(err)
		}

		return newMaxViewersActivator(r.deps, activator), nil
	case entity.ActivatorTypePercentage:
		activator, err := r.deps.rolloutService.FindPercentageActivatorByID(ctx, r.rollout.ActivatorID)
		if err != nil {
			r.deps.logger.ErrorWithContext(ctx, err)
			return nil, errs.ToResolverErr(err)
		}

		return newPercentageActivator(r.deps, activator), nil
	}

	return nil, nil
}

func (m Mutation) CreateRollout(
	ctx context.Context,
	args struct {
		SelectorType  entity.SelectorType
		ActivatorID   graphql.ID
		ActivatorType entity.ActivatorType
		IsEnabled     bool
	},
) (Rollout, error) {
	activatorID, internalErr := fromGraphQLID(args.ActivatorID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return Rollout{}, errs.ToResolverErr(internalErr)
	}

	rollout, err := m.deps.rolloutService.CreateRollout(
		ctx,
		args.SelectorType,
		activatorID,
		args.ActivatorType,
		args.IsEnabled,
	)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return Rollout{}, errs.ToResolverErr(err)
	}

	return newRollout(m.deps, rollout), nil
}

func (m Mutation) DeleteRollout(
	ctx context.Context,
	args struct {
		RolloutID graphql.ID
	},
) (Rollout, error) {
	rolloutID, internalErr := fromGraphQLID(args.RolloutID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return Rollout{}, errs.ToResolverErr(internalErr)
	}

	rollout, err := m.deps.rolloutService.DeleteRollout(ctx, rolloutID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return Rollout{}, errs.ToResolverErr(err)
	}

	return newRollout(m.deps, rollout), nil
}

func newRollout(deps *Dependencies, rollout entity.Rollout) Rollout {
	return Rollout{deps: deps, rollout: rollout}
}
