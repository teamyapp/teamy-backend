package entityv2

import "time"

type Message struct {
	ID        uint64
	Body      string
	ThreadID  uint64
	AuthorID  uint64
	CreatedAt time.Time
	UpdatedAt *time.Time
}
