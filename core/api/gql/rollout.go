package gql

import (
	"context"
	"fmt"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/service"
)

type Rollout struct {
	deps    *Dependencies
	rollout entity.Rollout
}

func (r Rollout) ID() graphql.ID {
	return toGraphQLID(r.rollout.ID)
}

func (r Rollout) IsEnabled() bool {
	return r.rollout.IsEnabled
}

func (r Rollout) VersionSelector(ctx context.Context) (VersionSelector, error) {
	versionSelector, err := r.deps.rolloutService.FindVersionSelectorByID(ctx, r.rollout.SelectorID)
	if err != nil {
		r.deps.logger.ErrorWithContext(ctx, err)
		return VersionSelector{}, errs.ToResolverErr(err)
	}

	return newVersionSelector(versionSelector, r.deps), nil
}

func (r Rollout) CreatedAt() graphql.Time {
	return toGraphQLTime(r.rollout.CreatedAt)
}

func (r Rollout) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(r.rollout.UpdatedAt)
}

func (r Rollout) Activator(ctx context.Context) (Activator, error) {
	activatorType, err := r.deps.rolloutService.FindActivatorTypeByID(ctx, r.rollout.ActivatorID)
	if err != nil {
		r.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

	switch activatorType {
	case entity.ActivatorTypeTimeRange:
		activator, err := r.deps.rolloutService.FindTimeRangeActivatorByID(ctx, r.rollout.ActivatorID)
		if err != nil {
			r.deps.logger.ErrorWithContext(ctx, err)
			return nil, errs.ToResolverErr(err)
		}

		return newTimeRangeActivator(activator), nil
	case entity.ActivatorTypeMaxViewers:
		activator, err := r.deps.rolloutService.FindMaxViewersActivatorByID(ctx, r.rollout.ActivatorID)
		if err != nil {
			r.deps.logger.ErrorWithContext(ctx, err)
			return nil, errs.ToResolverErr(err)
		}

		return newMaxViewersActivator(activator), nil
	case entity.ActivatorTypePercentage:
		activator, err := r.deps.rolloutService.FindPercentageActivatorByID(ctx, r.rollout.ActivatorID)
		if err != nil {
			r.deps.logger.ErrorWithContext(ctx, err)
			return nil, errs.ToResolverErr(err)
		}

		return newPercentageActivator(activator), nil
	}

	err = errs.NewError(
		errs.NotReady, fmt.Sprintf("activator type: %v is not ready", activatorType),
	)
	return nil, errs.ToResolverErr(err)
}

func (m Mutation) CreateAppRollout(
	ctx context.Context,
	args struct {
		AppID          graphql.ID
		AppRolloutType entity.AppRolloutRelationType
		Input          struct {
			VersionSelectorID graphql.ID
			ActivatorID       graphql.ID
			IsEnabled         bool
		}
	},
) (Rollout, error) {
	appID, internalErr := fromGraphQLID(args.AppID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return Rollout{}, errs.ToResolverErr(internalErr)
	}

	activatorID, internalErr := fromGraphQLID(args.Input.ActivatorID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return Rollout{}, errs.ToResolverErr(internalErr)
	}

	selectorID, internalErr := fromGraphQLID(args.Input.VersionSelectorID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return Rollout{}, errs.ToResolverErr(internalErr)
	}

	createAppRolloutInput := service.CreateRolloutInput{
		VersionSelectorID: selectorID,
		ActivatorID:       activatorID,
		IsEnabled:         args.Input.IsEnabled,
	}

	rollout, err := m.deps.rolloutService.CreateAppRollout(
		ctx,
		appID,
		args.AppRolloutType,
		createAppRolloutInput,
	)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return Rollout{}, errs.ToResolverErr(err)
	}

	return newRollout(m.deps, rollout), nil
}

func (m Mutation) UpdateRollout(
	ctx context.Context,
	args struct {
		RolloutID graphql.ID
		Input     struct {
			ActivatorID       graphql.ID
			VersionSelectorID graphql.ID
			IsEnabled         bool
		}
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

	activatorID, internalErr := fromGraphQLID(args.Input.ActivatorID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return Rollout{}, errs.ToResolverErr(internalErr)
	}

	selectorID, internalErr := fromGraphQLID(args.Input.VersionSelectorID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return Rollout{}, errs.ToResolverErr(internalErr)
	}

	updateRolloutInput := service.UpdateRolloutInput{
		VersionSelectorID: selectorID,
		ActivatorID:       activatorID,
		IsEnabled:         args.Input.IsEnabled,
	}

	rollout, err := m.deps.rolloutService.UpdateRollout(ctx, rolloutID, updateRolloutInput)
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
