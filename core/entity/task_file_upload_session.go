package entity

import (
	"time"
)

type TaskFileUploadSessionType string

const (
	TaskFileUploadSessionTypeContext TaskFileUploadSessionType = "Context"
)

type TaskFileUploadSession struct {
	TaskID              uint64
	Type                TaskFileUploadSessionType
	FileUploadSessionID uint64
	IsCompleted         bool
	CreatedAt           time.Time
	UpdatedAt           *time.Time
}
