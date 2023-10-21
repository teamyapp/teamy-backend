package entity

import "time"

type StoryTaskRelation struct {
	StoryID   uint64
	TaskID    uint64
	CreatedAt time.Time
}
