package resolver

import (
	"github.com/teamyapp/teamy-backend/app/entity"
)

type User struct {
	Entity
	deps *Dependencies
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

func (u User) ActiveTeam() (*Team, error) {
	teams := u.deps.Data.FilterTeams(func(t entity.Team) bool {
		return t.ID == u.user.ActiveTeamID
	})
	if len(teams) == 0 {
		return nil, nil
	}
	gqlTeam := newTeam(u.deps, teams[0])
	return &gqlTeam, nil
}

func (u User) Teams() ([]Team, error) {
	teams := u.deps.Data.FilterTeams(func(t entity.Team) bool {
		isCreator := t.CreatorID == u.user.ID
		if isCreator {
			return true
		}
		for _, id := range t.MemberIDs {
			if id == u.user.ID { // is member
				return true
			}
		}
		return false
	})
	return newTeams(u.deps, teams), nil
}

func (u User) Tasks(args struct{ Input *TaskFilter }) ([]Task, error) {
	q := NewQuery(u.deps)
	if args.Input != nil {
		id := u.ID()
		args.Input.CreatorID = &id
	}
	return q.Tasks(args)
}

func newUser(deps *Dependencies, user entity.User) User {
	return User{
		Entity: Entity{entity: user.Entity},
		deps:   deps,
		user:   user,
	}
}
