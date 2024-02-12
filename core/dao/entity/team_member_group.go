package entity

import "time"

type TeamMemberGroup struct {
	ID                       uint64
	Name                     string
	TeamID                   uint64
	AuthorizationUserGroupID uint64
	CreatedAt                time.Time
	UpdatedAt                *time.Time
}
