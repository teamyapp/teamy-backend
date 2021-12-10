package entity

import (
	"github.com/teamyapp/one/entity"
)

type Team struct {
	entity.Entity
	Name      string
	IconURL   *string
	CreatorID entity.ID
	MemberIDs OrderedSet_ID
	Tasks     OrderedSet_ID
	// user id -> task id
	NeedAttentionTasks map[entity.ID]entity.ID
}

type OrderedSet_ID []entity.ID

func (set OrderedSet_ID) Add(newID entity.ID) OrderedSet_ID {
	for _, id := range set {
		if id == newID {
			return set
		}
	}
	set = append(set, newID)
	return set
}
