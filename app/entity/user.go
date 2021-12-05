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
