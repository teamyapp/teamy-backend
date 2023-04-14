package entity

import (
	"time"
)

const OwnerTeamGroupLabel = "Owner"
const AdminTeamGroupLabel = "Admin"
const MemberTeamGroupLabel = "Member"

type TeamGroup struct {
	TeamID      uint64
	Label       string
	UserGroupID uint64
	CreatedAt   time.Time
}
