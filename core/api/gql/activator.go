package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/service"
)

type Activator interface {
	ID() graphql.ID
	Type() entity.ActivatorType
	CreatedAt() graphql.Time
	UpdatedAt() *graphql.Time
	ToStaticActivator() (*StaticActivator, bool)
	ToTimeRangeActivator() (*TimeRangeActivator, bool)
	ToMaxViewersActivator() (*MaxViewersActivator, bool)
	ToPercentageActivator() (*PercentageActivator, bool)
}

type StaticActivator struct {
	staticActivator entity.StaticActivator
}

var _ Activator = (*StaticActivator)(nil)

func (s StaticActivator) ID() graphql.ID {
	return toGraphQLID(s.staticActivator.ID)
}

func (s StaticActivator) Type() entity.ActivatorType {
	return entity.ActivatorTypeStatic
}

func (s StaticActivator) CreatedAt() graphql.Time {
	return toGraphQLTime(s.staticActivator.CreatedAt)
}

func (s StaticActivator) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(s.staticActivator.UpdatedAt)
}

func (s StaticActivator) ToStaticActivator() (*StaticActivator, bool) {
	return &s, true
}

func (s StaticActivator) ToTimeRangeActivator() (*TimeRangeActivator, bool) {
	return nil, false
}

func (s StaticActivator) ToMaxViewersActivator() (*MaxViewersActivator, bool) {
	return nil, false
}

func (s StaticActivator) ToPercentageActivator() (*PercentageActivator, bool) {
	return nil, false
}

func newStaticActivator(
	staticActivator entity.StaticActivator,
) StaticActivator {
	return StaticActivator{staticActivator: staticActivator}
}

type TimeRangeActivator struct {
	timeRangeActivator entity.TimeRangeActivator
}

var _ Activator = (*TimeRangeActivator)(nil)

func (t TimeRangeActivator) ID() graphql.ID {
	return toGraphQLID(t.timeRangeActivator.ID)
}

func (t TimeRangeActivator) Type() entity.ActivatorType {
	return entity.ActivatorTypeTimeRange
}

func (t TimeRangeActivator) StartAt() *graphql.Time {
	return toGraphQLTimePtr(t.timeRangeActivator.StartAt)
}

func (t TimeRangeActivator) EndAt() *graphql.Time {
	return toGraphQLTimePtr(t.timeRangeActivator.EndAt)
}

func (t TimeRangeActivator) CreatedAt() graphql.Time {
	return toGraphQLTime(t.timeRangeActivator.CreatedAt)
}

func (t TimeRangeActivator) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(t.timeRangeActivator.UpdatedAt)
}

func (t TimeRangeActivator) ToStaticActivator() (*StaticActivator, bool) {
	return nil, false
}

func (t TimeRangeActivator) ToTimeRangeActivator() (*TimeRangeActivator, bool) {
	return &t, true
}

func (t TimeRangeActivator) ToMaxViewersActivator() (*MaxViewersActivator, bool) {
	return nil, false
}

func (t TimeRangeActivator) ToPercentageActivator() (*PercentageActivator, bool) {
	return nil, false
}

func newTimeRangeActivator(
	timeRangeActivator entity.TimeRangeActivator,
) TimeRangeActivator {
	return TimeRangeActivator{timeRangeActivator: timeRangeActivator}
}

type MaxViewersActivator struct {
	maxViewersActivator entity.MaxViewersActivator
}

var _ Activator = (*MaxViewersActivator)(nil)

func (m MaxViewersActivator) ID() graphql.ID {
	return toGraphQLID(m.maxViewersActivator.ID)
}

func (m MaxViewersActivator) Type() entity.ActivatorType {
	return entity.ActivatorTypeMaxViewers
}

func (m MaxViewersActivator) MaxViewers() int32 {
	return int32(m.maxViewersActivator.MaxViewers)
}

func (m MaxViewersActivator) CreatedAt() graphql.Time {
	return toGraphQLTime(m.maxViewersActivator.CreatedAt)
}

func (m MaxViewersActivator) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(m.maxViewersActivator.UpdatedAt)
}

func (m MaxViewersActivator) ToStaticActivator() (*StaticActivator, bool) {
	return nil, false
}

func (m MaxViewersActivator) ToTimeRangeActivator() (*TimeRangeActivator, bool) {
	return nil, false
}

func (m MaxViewersActivator) ToMaxViewersActivator() (*MaxViewersActivator, bool) {
	return &m, true
}

func (m MaxViewersActivator) ToPercentageActivator() (*PercentageActivator, bool) {
	return nil, false
}

func newMaxViewersActivator(
	maxViewersActivator entity.MaxViewersActivator,
) MaxViewersActivator {
	return MaxViewersActivator{maxViewersActivator: maxViewersActivator}
}

type PercentageActivator struct {
	percentageActivator entity.PercentageActivator
}

var _ Activator = (*PercentageActivator)(nil)

func (p PercentageActivator) ID() graphql.ID {
	return toGraphQLID(p.percentageActivator.ID)
}

func (p PercentageActivator) Type() entity.ActivatorType {
	return entity.ActivatorTypePercentage
}

func (p PercentageActivator) Percentage() int32 {
	return int32(p.percentageActivator.Percentage)
}

func (p PercentageActivator) CreatedAt() graphql.Time {
	return toGraphQLTime(p.percentageActivator.CreatedAt)
}

func (p PercentageActivator) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(p.percentageActivator.UpdatedAt)
}

func (p PercentageActivator) ToStaticActivator() (*StaticActivator, bool) {
	return nil, false
}

func (p PercentageActivator) ToTimeRangeActivator() (*TimeRangeActivator, bool) {
	return nil, false
}

func (p PercentageActivator) ToMaxViewersActivator() (*MaxViewersActivator, bool) {
	return nil, false
}

func (p PercentageActivator) ToPercentageActivator() (*PercentageActivator, bool) {
	return &p, true
}

func newPercentageActivator(
	percentageActivator entity.PercentageActivator,
) PercentageActivator {
	return PercentageActivator{percentageActivator: percentageActivator}
}

func (m Mutation) CreateStaticActivator(
	ctx context.Context,
) (StaticActivator, error) {
	staticActivator, err := m.deps.rolloutService.CreateStaticActivator(ctx)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return StaticActivator{}, errs.ToResolverErr(err)
	}

	return newStaticActivator(staticActivator), nil
}

func (m Mutation) CreateTimeRangeActivator(
	ctx context.Context,
	args struct {
		Input struct {
			StartAt *graphql.Time
			EndAt   *graphql.Time
		}
	},
) (TimeRangeActivator, error) {
	startAt := fromGraphQLTimePtr(args.Input.StartAt)
	endAt := fromGraphQLTimePtr(args.Input.EndAt)
	timeRangeActivator, err := m.deps.rolloutService.CreateTimeRangeActivator(ctx, startAt, endAt)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return TimeRangeActivator{}, errs.ToResolverErr(err)
	}

	return newTimeRangeActivator(timeRangeActivator), nil
}

func (m Mutation) CreateMaxViewersActivator(
	ctx context.Context,
	args struct {
		Input struct {
			MaxViewers int32
		}
	},
) (MaxViewersActivator, error) {
	maxViewers := int(args.Input.MaxViewers)
	maxViewersActivator, err := m.deps.rolloutService.CreateMaxViewersActivator(ctx, maxViewers)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return MaxViewersActivator{}, errs.ToResolverErr(err)
	}

	return newMaxViewersActivator(maxViewersActivator), nil
}

func (m Mutation) CreatePercentageActivator(
	ctx context.Context,
	args struct {
		Input struct {
			Percentage int32
		}
	},
) (PercentageActivator, error) {
	percentage := int(args.Input.Percentage)
	percentageActivator, err := m.deps.rolloutService.CreatePercentageActivator(ctx, percentage)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return PercentageActivator{}, errs.ToResolverErr(err)
	}

	return newPercentageActivator(percentageActivator), nil
}

func (m Mutation) UpdateActivator(
	ctx context.Context,
	args struct {
		ActivatorID graphql.ID
		Input       struct {
			StartAt    *graphql.Time
			EndAt      *graphql.Time
			MaxViewers *int32
			Percentage *int32
			Type       entity.ActivatorType
		}
	},
) (Activator, error) {
	activatorID, internalErr := fromGraphQLID(args.ActivatorID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	var maxViewers int
	if args.Input.MaxViewers != nil {
		maxViewers = int(*args.Input.MaxViewers)
	}

	var percentage int
	if args.Input.Percentage != nil {
		percentage = int(*args.Input.Percentage)
	}

	updateActivatorInput := service.UpdateActivatorInput{
		Type:       args.Input.Type,
		StartAt:    fromGraphQLTimePtr(args.Input.StartAt),
		EndAt:      fromGraphQLTimePtr(args.Input.EndAt),
		MaxViewers: maxViewers,
		Percentage: percentage,
	}
	activator, err := m.deps.rolloutService.UpdateActivator(ctx, activatorID, updateActivatorInput)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

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
		return nil, errs.ToResolverErr(errs.NewError(errs.Unknown, "unknown activator type"))
	}

}

func (m Mutation) DeleteActivator(
	ctx context.Context,
	args struct {
		ActivatorID graphql.ID
	},
) (Activator, error) {
	activatorID, internalErr := fromGraphQLID(args.ActivatorID)
	if internalErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			internalErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ctx, internalErr)
		return nil, errs.ToResolverErr(internalErr)
	}

	activator, err := m.deps.rolloutService.DeleteActivator(ctx, activatorID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ctx, err)
		return nil, errs.ToResolverErr(err)
	}

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
		return nil, errs.ToResolverErr(errs.NewError(errs.Unknown, "unknown activator type"))
	}
}
