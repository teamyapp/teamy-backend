package entity

import "time"

type UserGroup struct {
	ID          uint64
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   *time.Time
}
