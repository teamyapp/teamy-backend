package entity

import "time"

type TaskLink struct {
	ID        uint64
	TaskID    uint64
	Title     string
	Url       string
	IconUrl   *string
	CreatedAt time.Time
	UpdatedAt *time.Time
}
