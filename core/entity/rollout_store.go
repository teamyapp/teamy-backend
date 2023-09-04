package entity

import "time"

type RolloutStore struct {
	RolloutID    uint64
	TotalViewers int
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}
