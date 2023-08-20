package entity

import (
	"time"
)

type AppPackageUploadSession struct {
	AppID               uint64
	UserID              uint64
	VersionNumber       int32
	FileUploadSessionID uint64
	IsCompleted         bool
	CreatedAt           time.Time
	UpdatedAt           *time.Time
}
