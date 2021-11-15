package service

import (
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
	"github.com/teamyapp/teamy-backend/app/repo"
)

type Team struct {
	teamRepo repo.Team
}

func (t Team) CreateTeam(ownerUserID oneEntity.ID, team entity.Team) error {
	panic("not implemented")
}

func (t Team) GetActiveTeam(userID oneEntity.ID) (entity.Team, error) {
	return t.teamRepo.GetActiveTeam(userID)
}

func (t Team) ListTeams(userID oneEntity.ID) ([]entity.Team, error) {
	panic("not implemented")
}

func (t Team) SetActiveTeam(userID oneEntity.ID, teamID oneEntity.ID) error {
	panic("not implemented")
}

func NewTeam(teamRepo repo.Team) Team {
	return Team{teamRepo: teamRepo}
}
