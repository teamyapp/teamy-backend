package gql

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Activator interface {
	ID() uint64
	CreatedAt() graphql.Time
	UpdatedAt() *graphql.Time
}

type TimeRangeActivator struct {
	deps               *Dependencies
	timeRangeActivator entity.TimeRangeActivator
}

var _ Activator = (*TimeRangeActivator)(nil)

func (t TimeRangeActivator) ID() uint64 {
	return t.timeRangeActivator.ID
}

func (t TimeRangeActivator) StartAt() graphql.Time {
	return toGraphQLTime(t.timeRangeActivator.StartAt)
}

func (t TimeRangeActivator) EndAt() graphql.Time {
	return toGraphQLTime(t.timeRangeActivator.EndAt)
}

func (t TimeRangeActivator) CreatedAt() graphql.Time {
	return toGraphQLTime(t.timeRangeActivator.CreatedAt)
}

func (t TimeRangeActivator) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(t.timeRangeActivator.UpdatedAt)
}

func newTimeRangeActivator(
	deps *Dependencies,
	timeRangeActivator entity.TimeRangeActivator,
) TimeRangeActivator {
	return TimeRangeActivator{deps: deps, timeRangeActivator: timeRangeActivator}
}

type MaxViewersActivator struct {
	deps                *Dependencies
	maxViewersActivator entity.MaxViewersActivator
}

var _ Activator = (*MaxViewersActivator)(nil)

func (m MaxViewersActivator) ID() uint64 {
	return m.maxViewersActivator.ID
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
	deps *Dependencies,
	maxViewersActivator entity.MaxViewersActivator,
) MaxViewersActivator {
	return MaxViewersActivator{deps: deps, maxViewersActivator: maxViewersActivator}
}

type PercentageActivator struct {
	deps                *Dependencies
	percentageActivator entity.PercentageActivator
}

var _ Activator = (*PercentageActivator)(nil)

func (p PercentageActivator) ID() uint64 {
	return p.percentageActivator.ID
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
	deps *Dependencies,
	percentageActivator entity.PercentageActivator,
) PercentageActivator {
	return PercentageActivator{deps: deps, percentageActivator: percentageActivator}
}
