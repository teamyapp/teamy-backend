package entity

import "time"

type ActivatorType string

const (
	ActivatorTypeTimeRange  ActivatorType = "TIME_RANGE"
	ActivatorTypeMaxViewers ActivatorType = "MAX_VIEWERS"
	ActivatorTypePercentage ActivatorType = "PERCENTAGE"
)

type TimeRangeActivator struct {
	ID        uint64
	StartAt   time.Time
	EndAt     time.Time
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type MaxViewersActivator struct {
	ID         uint64
	MaxViewers int
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}

type PercentageActivator struct {
	ID         uint64
	Percentage int
	CreatedAt  time.Time
	UpdatedAt  *time.Time
}
