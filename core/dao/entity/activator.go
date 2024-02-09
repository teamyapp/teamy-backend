package entity

import "time"

type PartialTimeRangeActivator struct {
	StartAt *time.Time
	EndAt   *time.Time
}

type PartialPercentageActivator struct {
	Percentage int
}

type PartialMaxViewersActivator struct {
	MaxViewers int
}
