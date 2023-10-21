package entity

import "time"

type StoryLink struct {
	ID        uint64
	StoryID   uint64
	Title     string
	URL       string
	IconURL   *string
	CreatedAt time.Time
	UpdatedAt *time.Time
}
