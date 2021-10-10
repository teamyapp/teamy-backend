package entity

import (
	"time"
)

type Entity struct {
	ID        int
	CreatedAt time.Time
	UpdatedAt *time.Time
}
