package entity

import (
	"time"

	"github.com/teamyapp/one/entity"
)

type Team struct {
	ID        entity.ID
	CreatedAt time.Time
	UpdatedAt *time.Time
	Name      string
	IconURL   string
	CreatorID entity.ID
	MemberIDs OrderedSetID
	Tasks     OrderedSetID
	// user id -> task id
	NeedAttentionTasks map[entity.ID]entity.ID
	InvitationIDs      OrderedSetID
}

// OrderedSetID will be refactored once Go 1.18 is released with Generics support.
type OrderedSetID []entity.ID

func (set OrderedSetID) Add(newID entity.ID) OrderedSetID {
	for _, id := range set {
		if id == newID {
			return set
		}
	}
	set = append(set, newID)
	return set
}

func (set OrderedSetID) Has(id entity.ID) bool {
	for _, ID := range set {
		if ID == id {
			return true
		}
	}
	return false
}

func (set OrderedSetID) Remove(id entity.ID) OrderedSetID {
	for i, ID := range set {
		if ID == id {
			return append(set[:i], set[i+1:]...)
		}
	}
	return set
}
