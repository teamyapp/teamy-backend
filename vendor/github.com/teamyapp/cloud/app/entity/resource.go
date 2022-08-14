package entity

import "time"

type Resource struct {
	ResourceTypeName string
	ResourceID       uint64
	CreatedAt        time.Time
	CreatorUserID    uint64
}
