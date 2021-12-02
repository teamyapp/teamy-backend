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
	team, err := u.deps.teamRepo.FindActiveTeam(u.user.ID)
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

func (u User) Teams() ([]Team, error) {
	teamIDs, err := u.deps.teamRepo.FindAllTeamIDs(u.user.ID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	teams, err := u.deps.teamRepo.FindTeams(teamIDs)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	var gqlTeams = make([]Team, 0)
	for _, team := range teams {
		gqlTeam := newTeam(u.deps, u.prototypeDeps, team)
		gqlTeams = append(gqlTeams, gqlTeam)
	}
	return gqlTeams, nil
}

func newUser(deps *Dependencies, prototypeDeps *resolver.Dependencies, user entity.User) User {
	return User{
		Entity:        Entity{entity: user.Entity},
		deps:          deps,
		prototypeDeps: prototypeDeps,
		user:          user,
	}
}
