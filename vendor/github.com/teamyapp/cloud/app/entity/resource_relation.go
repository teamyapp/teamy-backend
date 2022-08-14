package entity

import "time"

type ResourceRelation struct {
	ChildResourceID    uint64
	ChildResourceType  string
	ParentResourceID   uint64
	ParentResourceType string
	CreatedAt          time.Time
	CreatorUserID      uint64
}
