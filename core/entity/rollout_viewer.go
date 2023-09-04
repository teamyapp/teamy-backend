package entity

import "time"

type RolloutViewer struct {
	RolloutID     uint64
	ViewerID      uint64
	VersionNumber int
	IsActivated   bool
	CreatedAt     time.Time
	UpdatedAt     *time.Time
}
