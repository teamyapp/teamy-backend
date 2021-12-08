package entity

import (
	"github.com/teamyapp/one/entity"
)

type Team struct {
	entity.Entity
	Name               string
	IconURL            *string
	CreatorID          entity.ID
	MemberIDs          []entity.ID
	Tasks              []entity.ID
	NeedAttentionTasks map[entity.ID]entity.ID
}

type TeamUserTaskNeedAttention struct {
	UserID entity.ID
	TaskID entity.ID
}
