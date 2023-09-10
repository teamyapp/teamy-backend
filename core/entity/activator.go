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

type ActivatorUnion struct {
	Type                ActivatorType
	TimeRangeActivator  TimeRangeActivator
	MaxViewersActivator MaxViewersActivator
	PercentageActivator PercentageActivator
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

type ActivatorTypeRelation struct {
	ActivatorID   uint64
	ActivatorType ActivatorType
}
