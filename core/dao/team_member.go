package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamMember interface {
	FindTeamIDsByUserID(ct context.Context, userID uint64) ([]uint64, error)
	FindTeamMemberIDsByTeamID(ct context.Context, teamID uint64) ([]uint64, error)
	FindTeamMembersByTeamID(ct context.Context, teamID uint64) ([]entity.TeamMember, error)
	FindTeamMember(ct context.Context, teamID uint64, userID uint64) (entity.TeamMember, error)
	CreateTeamMember(ct context.Context, teamMember entity.TeamMember) error
	UpdateTeamMember(ct context.Context, teamMember entity.TeamMember) error
	DeleteTeamMember(ct context.Context, teamID uint64, userID uint64) error
}
