package dao

import (
	"context"
)

type TeamMember interface {
	FindTeamIDsByUserID(ct context.Context, userID uint64) ([]uint64, error)
	FindTeamMemberIDsByTeamID(ct context.Context, teamID uint64) ([]uint64, error)
	HasTeamMember(ct context.Context, teamID uint64, userID uint64) (bool, error)
	CreateTeamMember(ct context.Context, teamID uint64, userID uint64) error
	DeleteTeamMember(ct context.Context, teamID uint64, userID uint64) error
}
