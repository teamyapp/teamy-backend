package service

import (
	"log"

	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
	"github.com/teamyapp/teamy-backend/app/repo"
)

type User struct {
	userRepo repo.User
	teamRepo repo.Team
}

func (u User) FindUser(userID oneEntity.ID) (entity.User, error) {
	user, err := u.userRepo.GetUser(userID)
	if err != nil {
		log.Println(err)
		return entity.User{}, err
	}

	return user, nil
}

func NewUser(userRepo repo.User, teamRepo repo.Team) User {
	return User{
		userRepo: userRepo,
		teamRepo: teamRepo,
	}
}
