package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Activator interface {
	ID() graphql.ID
	CreatedAt() graphql.Time
	UpdatedAt() *graphql.Time
}

type TimeRangeActivator struct {
	timeRangeActivator entity.TimeRangeActivator
}

var _ Activator = (*TimeRangeActivator)(nil)

func (t TimeRangeActivator) ID() graphql.ID {
	return toGraphQLID(t.timeRangeActivator.ID)
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

func (m MaxViewersActivator) MaxViewers() int32 {
	return int32(m.maxViewersActivator.MaxViewers)
}

func (m MaxViewersActivator) CreatedAt() graphql.Time {
	return toGraphQLTime(m.maxViewersActivator.CreatedAt)
}

func (m MaxViewersActivator) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(m.maxViewersActivator.UpdatedAt)
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

func (p PercentageActivator) Percentage() int32 {
	return int32(p.percentageActivator.Percentage)
}

func (p PercentageActivator) CreatedAt() graphql.Time {
	return toGraphQLTime(p.percentageActivator.CreatedAt)
}

func (p PercentageActivator) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(p.percentageActivator.UpdatedAt)
}

func newPercentageActivator(
	percentageActivator entity.PercentageActivator,
) PercentageActivator {
	return PercentageActivator{percentageActivator: percentageActivator}
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
