package entity

import "time"

type SprintStoryRelation struct {
	SprintID  uint64
	StoryID   uint64
	CreatedAt time.Time
}
