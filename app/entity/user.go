package entity

import (
	"time"
)

type User struct {
	ID           uint64
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	FirstName    string
	LastName     string
	ProfileURL   string
	ActiveTeamID uint64
}

func GhostUser() User {
	return User{
		ID:        0,
		FirstName: "Ghost",
		LastName:  "this user has either been removed from the system or never existed",
	}
}
