package entity

import (
	"time"
)

type TeamFileUploadSessionType string

const (
	IconTeamFileUploadSessionType TeamFileUploadSessionType = "ICON"
)

type TeamFileUploadSession struct {
	TeamID              uint64
	Type                TeamFileUploadSessionType
	FileUploadSessionID uint64
	IsCompleted         bool
	CreatedAt           time.Time
	UpdatedAt           *time.Time
}
