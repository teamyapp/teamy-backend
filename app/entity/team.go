package entity

import (
	oneEntity "github.com/teamyapp/one/entity"
)

type Team struct {
	oneEntity.Entity
	Name          string
	LogoURL       string
	MemberUserIds []int
}
