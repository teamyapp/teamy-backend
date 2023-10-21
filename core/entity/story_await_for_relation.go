package entity

import "time"

type StoryAwaitForRelation struct {
	AwaitingStoryID uint64
	AwaitForStoryID uint64
	CreatedAt       time.Time
}
