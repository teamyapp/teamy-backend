package service

import (
	"github.com/teamyapp/teamy-backend/app/repo"
)

type User struct {
	userRepo repo.User
	teamRepo repo.Team
}

func NewUser(userRepo repo.User, teamRepo repo.Team) User {
	return User{
		userRepo: userRepo,
		teamRepo: teamRepo,
	}
}
