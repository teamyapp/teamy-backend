package service

import (
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
	"github.com/teamyapp/teamy-backend/app/repo"
	"log"
)

type User struct {
	userRepo repo.User
	teamRepo repo.Team
}

func (u User) ListTeamMembers(teamID oneEntity.ID) ([]entity.User, error) {
	ids, err := u.teamRepo.ListTeamMemberIDs(teamID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	members, err := u.userRepo.GetUsers(ids)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return members, nil
}

func NewUser(userRepo repo.User, teamRepo repo.Team) User {
	return User{
		userRepo: userRepo,
		teamRepo: teamRepo,
	}
}


