package entity

import (
	"time"
)

type Entity struct {
	ID        ID
	CreatedAt time.Time
	UpdatedAt *time.Time
}
