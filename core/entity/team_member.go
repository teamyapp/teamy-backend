package entity

import (
	"time"
)

type TeamMember struct {
	TeamID          uint64
	UserID          uint64
	WeeklyBandwidth time.Duration
	CreatedAt       time.Time
	UpdatedAt       *time.Time
}
