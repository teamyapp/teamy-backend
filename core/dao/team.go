package dao

import (
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Team interface {
	FindAllTeams() ([]entity.Team, error)
	FindTeamByID(teamID uint64) (entity.Team, error)
	FindTeamsByIDs(teamIDs []uint64) ([]entity.Team, error)
	CreateTeam(team entity.Team) error
	UpdateTeam(team entity.Team) error
}
