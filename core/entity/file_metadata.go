package entity

import (
	"time"
)

type FileMetadata struct {
	ID             uint64
	Name           string
	OwningTeamID   uint64
	CreatedAt      time.Time
	LastModifiedAt *time.Time
}
