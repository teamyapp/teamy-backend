package entity

import (
	oneEntity "github.com/teamyapp/one/entity"
)

type User struct {
	oneEntity.Entity
	FirstName    string
	LastName     string
	ProfileURL   string
	ActiveTeamID oneEntity.ID
}

func GhostUser() User {
	return User{
		Entity: oneEntity.Entity{
			ID: -1,
		},
		FirstName: "Ghost",
		LastName: "this user has either been removed from the system or never existed",
	}
}
