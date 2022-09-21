package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/core/entity"
)

type Team interface {
	FindAllTeams(ct context.Context) ([]entity.Team, error)
	FindTeamByID(ct context.Context, teamID uint64) (entity.Team, error)
	FindTeamsByIDs(ct context.Context, teamIDs []uint64) ([]entity.Team, error)
	CreateTeam(ct context.Context, team entity.Team) error
	UpdateTeam(ct context.Context, team entity.Team) error
}
