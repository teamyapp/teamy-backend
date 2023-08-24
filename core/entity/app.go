package entity

import "time"

type App struct {
	ID                 uint64
	TotalInstallations int
	CreatedAt          time.Time
	UpdatedAt          *time.Time
	ManagedByTeamID    uint64
}
