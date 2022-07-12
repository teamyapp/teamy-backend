package entity

import "time"

type Permission struct {
	ResourceType string
	ResourceID   uint64
	Operation    string
	GroupID      uint64
	CreatedAt    time.Time
	UpdatedAt    *time.Time
}
