package entity

import (
	"github.com/teamyapp/one/entity"
)

type Team struct {
	entity.Entity
	Name      string
	LogoURL   *string
	CreatorID entity.ID
	MemberIDs []entity.ID
	Tasks     []entity.ID
}
