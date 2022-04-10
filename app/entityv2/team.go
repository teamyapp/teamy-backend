package entityv2

import (
	"time"
)

type Team struct {
	ID        uint64
	Name      string
	IconURL   string
	CreatorID uint64
	CreatedAt time.Time
	UpdatedAt *time.Time
}
