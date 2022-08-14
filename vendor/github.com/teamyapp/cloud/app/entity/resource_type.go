package entity

import "time"

type ResourceType struct {
	ResourceTypeName string
	CreatedAt        time.Time
	CreatorUserID    uint64
}
