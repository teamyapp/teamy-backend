package resolver

import "github.com/teamyapp/teamy-backend/app/entity"

type User struct {
	Entity
	user entity.User
}

func (u User) FirstName() string {
	// TODO: replace with real first name
	return u.user.FirstName
}

func (u User) LastName() string {
	// TODO: replace with real last name
	return u.user.LastName
}

func (u User) ProfileURL() string {
	return u.user.ProfileURL
}

func (u User) ActiveTeam() *Team {
	// todo: implement it
	return nil
}

func newUser(user entity.User) User {
	return User{
		Entity: Entity{entity: user.Entity},
		user:   user,
	}
}
