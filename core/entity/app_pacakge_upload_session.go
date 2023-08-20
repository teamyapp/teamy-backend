package entity

import (
	"time"
)

type AppPackageUploadSession struct {
	AppID               uint64
	VersionNumber       int
	FileUploadSessionID uint64
	IsCompleted         bool
	CreatedAt           time.Time
	UpdatedAt           *time.Time
}
