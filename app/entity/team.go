package entity

import (
	"time"
)

type Team struct {
	ID        uint64
	CreatedAt time.Time
	UpdatedAt *time.Time
	Name      string
	IconURL   string
	CreatorID uint64
	MemberIDs OrderedSetID
	Tasks     OrderedSetID
	// user id -> task id
	NeedAttentionTasks map[uint64]uint64
	InvitationIDs      OrderedSetID
}

// OrderedSetID will be refactored once Go 1.18 is released with Generics support.
type OrderedSetID []uint64

func (set OrderedSetID) Add(newID uint64) OrderedSetID {
	for _, id := range set {
		if id == newID {
			return set
		}
	}
	set = append(set, newID)
	return set
}

func (set OrderedSetID) Has(id uint64) bool {
	for _, ID := range set {
		if ID == id {
			return true
		}
	}
	return false
}

func (set OrderedSetID) Remove(id uint64) OrderedSetID {
	for i, ID := range set {
		if ID == id {
			return append(set[:i], set[i+1:]...)
		}
	}
	return set
}
