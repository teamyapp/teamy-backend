package entity

type RolloutType string

const (
	RolloutTypeStatic     RolloutType = "STATIC"
	RolloutTypeTimeRange  RolloutType = "TIME_RANGE"
	RolloutTypeExperiment RolloutType = "EXPERIMENT"
)
