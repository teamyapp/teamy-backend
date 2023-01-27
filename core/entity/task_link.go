package entity

import "time"

type TaskLink struct {
	ID        uint64
	TaskID    uint64
	Title     string
	URL       string
	IconURL   *string
	CreatedAt time.Time
	UpdatedAt *time.Time
}
