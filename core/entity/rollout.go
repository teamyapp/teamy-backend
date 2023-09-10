package entity

import (
	"time"
)

type Rollout struct {
	ID          uint64
	ActivatorID uint64
	SelectorID  uint64
	IsEnabled   bool
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}
