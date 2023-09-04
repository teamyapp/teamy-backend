package entity

import (
	"time"
)

type SelectorType string

const (
	SelectorTypeStatic     SelectorType = "STATIC"
	SelectorTypeExperiment SelectorType = "EXPERIMENT"
)

type Rollout struct {
	ID            uint64
	ActivatorID   uint64
	SelectorID    uint64
	ActivatorType ActivatorType
	SelectorType  SelectorType
	IsEnabled     bool
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}
