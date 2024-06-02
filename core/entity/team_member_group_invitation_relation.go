package entity

import "time"

type TeamMemberGroupInvitationRelation struct {
	GroupID      uint64
	InvitationID uint64
	CreatedAt    time.Time
}
