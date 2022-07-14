package entity

import (
	"time"
)

type Sprint struct {
	ID           uint64
	StartAt      time.Time
	EndAt        time.Time
	CreatedAt    time.Time
	OwningTeamID uint64
}
