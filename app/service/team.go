package service

import (
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Team struct {
}

func (t Team) CreateTeam(ownerUserID int, team entity.Team) error {
	panic("not implemented")
}

func (t Team) GetActiveTeam(userID int) (entity.Team, error) {
	panic("not implemented")
}

func (t Team) ListTeams(userID int) ([]entity.Team, error) {
	panic("not implemented")
}

func (t Team) SwitchActiveTeam(userID int, teamID int) error {
	panic("not implemented")
}
