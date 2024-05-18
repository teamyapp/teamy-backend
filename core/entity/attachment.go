package entity

import "time"

type AttachmentType string

const (
	AttachmentTypeImage AttachmentType = "IMAGE"
)

type Attachment struct {
	ID               uint64
	Type             AttachmentType
	URL              string
	Size             uint64
	AttachmentListID uint64
	CreatedAt        time.Time
	UpdatedAt        *time.Time
}
