package entityv2

import "time"

type Message struct {
	ID        uint64
	Body      string
	CreatedAt time.Time
	UpdatedAt *time.Time
}
