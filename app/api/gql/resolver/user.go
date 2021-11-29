package resolver

import "github.com/teamyapp/teamy-backend/app/entity"

type User struct {
	Entity
	user entity.User
}

func (u User) Name() string {
	return u.user.Name
}

func (u User) ProfileURL() string {
	return u.user.ProfileURL
}

func newUser(user entity.User) User {
	return User{
		Entity: Entity{entity: user.Entity},
		user:   user,
	}
}
