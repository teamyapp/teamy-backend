package entity

import "time"

type ActivatorType string

const (
	ActivatorTypeStatic     ActivatorType = "STATIC"
	ActivatorTypeTimeRange  ActivatorType = "TIME_RANGE"
	ActivatorTypeMaxViewers ActivatorType = "MAX_VIEWERS"
	ActivatorTypePercentage ActivatorType = "PERCENTAGE"
)

type Activator struct {
	ID        uint64
	Type      ActivatorType
	Locked    bool
	CreatedAt time.Time
	UpdatedAt *time.Time
}

type ActivatorUnion struct {
	Type                ActivatorType
	TimeRangeActivator  TimeRangeActivator
	MaxViewersActivator MaxViewersActivator
	PercentageActivator PercentageActivator
	StaticActivator     StaticActivator
}

type StaticActivator struct {
	Activator
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
