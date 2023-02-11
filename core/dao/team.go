package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Team interface {
	FindAllTeams(ct context.Context) ([]entity.Team, *errs.Error)
	FindTeamByID(ct context.Context, teamID uint64) (entity.Team, *errs.Error)
	FindTeamsByIDs(ct context.Context, teamIDs []uint64) ([]entity.Team, *errs.Error)
	CreateTeam(ct context.Context, team entity.Team) *errs.Error
	UpdateTeam(ct context.Context, team entity.Team) *errs.Error
	DeleteTeam(ct context.Context, teamID uint64) *errs.Error
}
