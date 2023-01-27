package entity

import (
	"time"
)

type SprintParticipant struct {
	SprintID        uint64
	UserID          uint64
	TotalBandwidth  time.Duration
	UnusedBandwidth time.Duration
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}
