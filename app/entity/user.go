package entity

import (
	"time"
)

type User struct {
	ID         uint64
	CreatedAt  time.Time
	UpdatedAt  *time.Time
	FirstName  string
	LastName   string
	ProfileURL *string
}
