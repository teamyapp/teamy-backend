package dao

import "github.com/teamyapp/teamy-backend/app/entityv2"

type Team interface {
	FindAllTeams() ([]entityv2.Team, error)
	FindTeamByID(teamID uint64) (entityv2.Team, error)
	FindTeamsByIDs(teamIDs []uint64) ([]entityv2.Team, error)
	CreateTeam(team entityv2.Team) error
	UpdateTeam(team entityv2.Team) error
}
