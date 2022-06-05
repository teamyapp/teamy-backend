package entity

import (
	"time"
)

type Team struct {
	ID            uint64
	Name          string
	IconURL       *string
	CreatorUserID uint64
	OwnerUserID   uint64
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}
