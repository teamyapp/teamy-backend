package entityv2

import (
	"time"
)

type Task struct {
	ID               uint64
	CreatedAt        time.Time
	UpdatedAt        *time.Time
	Goal             string
	DueAt            *time.Time
	Context          string
	CreatorID        uint64
	OwnerUserID      *uint64
	OwningTeamID     uint64
	Status           TaskStatus
	CommentsThreadID *uint64
}
