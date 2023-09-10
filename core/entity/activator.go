package entity

import "time"

type ActivatorType string

const (
	ActivatorTypeTimeRange  ActivatorType = "TIME_RANGE"
	ActivatorTypeMaxViewers ActivatorType = "MAX_VIEWERS"
	ActivatorTypePercentage ActivatorType = "PERCENTAGE"
)

type Activator struct {
	ID        uint64
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type ActivatorTypeRelation struct {
	ID            uint64
	ActivatorType ActivatorType
}

type TimeRangeActivator struct {
	Activator
	StartAt *time.Time
	EndAt   *time.Time
}

type MaxViewersActivator struct {
	Activator
	MaxViewers int
}

type PercentageActivator struct {
	Activator
	Percentage int
}
