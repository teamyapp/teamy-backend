package entityv2

import (
	"time"
)

type Task struct {
	ID           uint64
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	Goal         string
	DueAt        *time.Time
	Context      string
	CreatorID    uint64
	OwnerUserId  *uint64
	OwningTeamId uint64
	Status       TaskStatus
	Comments     *Thread
}
