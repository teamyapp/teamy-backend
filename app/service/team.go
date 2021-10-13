package service

import (
	"github.com/teamyapp/teamy-backend/app/entity"
	"github.com/teamyapp/teamy-backend/app/repo"
)

type Team struct {
	teamRepo repo.Team
}

func (t Team) CreateTeam(ownerUserID entity.ID, team entity.Team) error {
	panic("not implemented")
}

func (t Team) GetActiveTeam(userID entity.ID) (entity.Team, error) {
	return t.teamRepo.GetActiveTeam(userID)
}

func (t Team) ListTeams(userID entity.ID) ([]entity.Team, error) {
	panic("not implemented")
}

func (t Team) SetActiveTeam(userID entity.ID, teamID entity.ID) error {
	panic("not implemented")
}

func NewTeam(teamRepo repo.Team) Team {
	return Team{teamRepo: teamRepo}
}
