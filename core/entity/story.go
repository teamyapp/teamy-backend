package entity

import "time"

type StoryStatus string

const (
	TodoStoryStatus       StoryStatus = "TODO"
	InProgressStoryStatus StoryStatus = "IN_PROGRESS"
	PausedStoryStatus     StoryStatus = "PAUSED"
	CompletedStoryStatus  StoryStatus = "COMPLETED"
)

type Story struct {
	ID        uint64
	Name      string
	OwnerID   *uint64
	Status    StoryStatus
	Priority  *Priority
	IsPlanned bool
	CreatorID uint64
	CreatedAt time.Time
	UpdatedAt *time.Time
}
