package entity

import "time"

type Message struct {
	ID           uint64
	Body         string
	ThreadID     uint64
	AuthorUserID uint64
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}
