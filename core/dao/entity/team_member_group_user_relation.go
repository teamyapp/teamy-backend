package entity

import "time"

type TeamMemberGroupUserRelation struct {
	GroupID   uint64
	UserID    uint64
	CreatedAt time.Time
}
