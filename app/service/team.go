package service

import (
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
	"github.com/teamyapp/teamy-backend/app/repo"
	"log"
)

type Team struct {
	teamRepo repo.Team
	userRepo repo.User
}

func (t Team) CreateTeam(ownerUserID oneEntity.ID, team entity.Team) error {
	panic("not implemented")
}

func (t Team) GetActiveTeam(userID oneEntity.ID) (*entity.Team, error) {
	return t.teamRepo.GetActiveTeam(userID)
}

func (t Team) ListTeams(userID oneEntity.ID) ([]entity.Team, error) {
	panic("not implemented")
}

func (t Team) SetActiveTeam(userID oneEntity.ID, teamID oneEntity.ID) error {
	panic("not implemented")
}

func (t Team) ListTeamMembers(teamID oneEntity.ID) ([]entity.User, error) {
	ids, err := t.teamRepo.ListTeamMemberIDs(teamID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	members, err := t.userRepo.GetUsers(ids)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return members, nil
}

func NewTeam(teamRepo repo.Team, userRepo repo.User) Team {
	return Team{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}
