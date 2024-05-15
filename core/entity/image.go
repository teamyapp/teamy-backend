package entity

import "time"

type Image struct {
	ID               uint64
	URL              string
	Size             uint64
	AttachmentListID uint64
	CreatedAt        time.Time
	UpdatedAt        *time.Time
}
