package entity

import (
	"time"
)

type Rollout struct {
	ID          uint64
	Name        string
	ActivatorID uint64
	SelectorID  uint64
	Viewers     int
	IsEnabled   bool
	Locked      bool
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}
