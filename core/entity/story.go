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
	Priority  Priority
	CreatedAt time.Time
	UpdatedAt *time.Time
}
