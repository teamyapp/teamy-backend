package entity

import (
	"time"
)

type RolloutType string

const (
	RolloutTypeStatic     RolloutType = "STATIC"
	RolloutTypeTimeRange  RolloutType = "TIME_RANGE"
	RolloutTypeExperiment RolloutType = "EXPERIMENT"
)

type Rollout struct {
	ID        uint64
	Type      RolloutType
	CreatedAt time.Time
}

type TimeRangeRollout struct {
	Rollout
	StartAt *time.Time
	EndAt   *time.Time
}
