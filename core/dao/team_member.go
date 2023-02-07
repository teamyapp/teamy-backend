package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamMember interface {
	FindTeamIDsByUserID(ct context.Context, userID uint64) ([]uint64, *errs.Error)
	FindTeamMemberIDsByTeamID(ct context.Context, teamID uint64) ([]uint64, *errs.Error)
	FindTeamMembersByTeamID(ct context.Context, teamID uint64) ([]entity.TeamMember, *errs.Error)
	FindTeamMember(ct context.Context, teamID uint64, userID uint64) (entity.TeamMember, *errs.Error)
	CreateTeamMember(ct context.Context, teamMember entity.TeamMember) *errs.Error
	UpdateTeamMember(ct context.Context, teamMember entity.TeamMember) *errs.Error
	DeleteTeamMember(ct context.Context, teamID uint64, userID uint64) *errs.Error
}
