package entity

import "time"

type App struct {
	ID                 uint64
	TotalInstallations int
	ManagedByTeamID    uint64
	CreatedAt          time.Time
	UpdatedAt          *time.Time
}
