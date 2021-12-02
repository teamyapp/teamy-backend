package resolver

import (
	"log"

	"github.com/teamyapp/teamy-backend/app/api/gqlv2/resolver"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type User struct {
	Entity
	deps          *Dependencies
	prototypeDeps *resolver.Dependencies
	user          entity.User
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

func (u User) ActiveTeam() (*Team, error) {
	team, err := u.deps.executionService.GetActiveTeam(u.user.ID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if team == nil {
		return nil, nil
	}

	gqlTeam := newTeam(u.deps, u.prototypeDeps, *team)
	return &gqlTeam, nil
}

func newUser(deps *Dependencies, prototypeDeps *resolver.Dependencies, user entity.User) User {
	return User{
		Entity:        Entity{entity: user.Entity},
		deps:          deps,
		prototypeDeps: prototypeDeps,
		user:          user,
	}
}
