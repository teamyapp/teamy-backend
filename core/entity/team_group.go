package entity

import (
	"time"
)

const OwnerTeamGroupLabel = "OWNER"
const AdminTeamGroupLabel = "ADMIN"
const MemberTeamGroupLabel = "MEMBER"

type TeamGroup struct {
	TeamID      uint64
	Label       string
	UserGroupID uint64
	CreatedAt   time.Time
}
