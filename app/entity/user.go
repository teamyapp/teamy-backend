package entity

import (
	oneEntity "github.com/teamyapp/one/entity"
)

type User struct {
	oneEntity.Entity
	Name       string
	ProfileURL string
}
