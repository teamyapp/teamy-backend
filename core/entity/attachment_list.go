package entity

import "time"

type AttachmentListOwnerType string

const (
	AttachmentListOwnerTypeTask AttachmentListOwnerType = "TASK"
)

type AttachmentList struct {
	OwnerType AttachmentListOwnerType
	OwnerID   uint64
	ListLabel string
	ListID    uint64
	CreatedAt time.Time
	UpdatedAt *time.Time
}
