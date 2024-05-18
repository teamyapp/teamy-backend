package entity

import (
	"time"
)

type AttachmentFileUploadSession struct {
	AttachmentListID    uint64
	FileUploadSessionID uint64
	IsCompleted         bool
	CreatedAt           time.Time
	UpdatedAt           *time.Time
}
