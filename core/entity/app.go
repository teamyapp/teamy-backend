package entity

import "time"

type App struct {
	ID                  uint64
        Name             string
        Description         string
	APISecret           string
	ActiveVersionNumber *int32
	InstallationCount   uint64
	CreatorUserID       uint64
	CreatedAt           time.Time
	UpdatedAt           *time.Time
}
