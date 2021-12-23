package entity

import (
	"github.com/teamyapp/one/entity"
)

type Team struct {
	entity.Entity
	Name      string
	IconURL   string
	CreatorID entity.ID
	MemberIDs OrderedSet_ID
	Tasks     OrderedSet_ID
	// user id -> task id
	NeedAttentionTasks map[entity.ID]entity.ID
}

// The follow line doesn't follow Go's naming convention and will be refactored once Go 1.18 is released with Generics support.
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

func (set OrderedSet_ID) Has(id entity.ID) bool {
	for _, ID := range set {
		if ID == id {
			return true
		}
	}
	return false
}
