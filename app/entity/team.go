package entity

import (
	"github.com/teamyapp/one/entity"
)

type Team struct {
	entity.Entity
	Name          string
	LogoURL       string
	MemberIDs []entity.ID
}
