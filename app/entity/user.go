package entity

import (
	"time"

	oneEntity "github.com/teamyapp/one/entity"
)

type User struct {
	ID           oneEntity.ID
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	FirstName    string
	LastName     string
	ProfileURL   string
	ActiveTeamID oneEntity.ID
}

func GhostUser() User {
	return User{
		ID:        -1,
		FirstName: "Ghost",
		LastName:  "this user has either been removed from the system or never existed",
	}
}
