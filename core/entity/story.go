package entity

import "time"

type Story struct {
	StoryID          uint64
	Goal             string
	DueAt            *time.Time
	Context          *string
	CreatorUserID    uint64
	OwnerUserID      *uint64
	OwningTeamID     uint64
	Status           StoryStatus
	IsPlanned        bool
	Effort           *time.Duration
	CommentsThreadID uint64
	CreatedAt        time.Time
	UpdatedAt        *time.Time
	DeliveredAt      *time.Time
	Priority         *Priority
}
