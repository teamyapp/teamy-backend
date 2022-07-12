package entity

import "time"

type UserGroupMember struct {
	GroupID   uint64
	UserID    uint64
	CreatedAt time.Time
}
