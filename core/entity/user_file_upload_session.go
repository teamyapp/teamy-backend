package entity

import (
	"time"
)

type UserFileUploadSessionType string

const (
	ProfileUserFileUploadSessionType UserFileUploadSessionType = "PROFILE"
)

type UserFileUploadSession struct {
	UserID              uint64
	Type                UserFileUploadSessionType
	FileUploadSessionID uint64
	IsCompleted         bool
	CreatedAt           time.Time
	UpdatedAt           *time.Time
}
