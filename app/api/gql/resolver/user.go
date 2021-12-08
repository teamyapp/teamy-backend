package resolver

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type User struct {
	Entity
	deps *Dependencies
	user entity.User
}

type UserInput struct {
	ID         *graphql.ID
	FirstName  *string
	LastName   *string
	ProfileUrl *string
}

// Admin && Debug Only
func (m Mutation) CreateUser(
	args struct {
		Input struct {
			ID         graphql.ID
			FirstName  *string
			LastName   *string
			ProfileUrl *string
		}
	},
) (User, error) {
	id, err := fromGraphQLID(args.Input.ID)
	if err != nil {
		return User{}, err
	}
	firstName := ""
	if args.Input.FirstName != nil {
		firstName = *args.Input.FirstName
	}
	lastName := ""
	if args.Input.LastName != nil {
		lastName = *args.Input.LastName
	}
	profileURL := ""
	if args.Input.ProfileUrl != nil {
		profileURL = *args.Input.ProfileUrl
	}
	user, err := m.deps.Data.CreateUser(entity.User{
		Entity: oneEntity.Entity{
			ID: id,
		},
		FirstName:  firstName,
		LastName:   lastName,
		ProfileURL: profileURL,
	})
	if err != nil {
		return User{}, err
	}
	return newUser(m.deps, user), nil
}

func (m Mutation) UpdateUser(args struct{ Input UserInput }) (User, error) {
	if args.Input.ID == nil {
		return User{}, errors.New("must provide an ID")
	}
	id, err := fromGraphQLID(*args.Input.ID)
	if err != nil {
		return User{}, err
	}
	user, err := m.deps.Data.UpdateUser(id, func(u entity.User) entity.User {
		if args.Input.LastName != nil {
			u.LastName = *args.Input.LastName
		}
		if args.Input.FirstName != nil {
			u.FirstName = *args.Input.FirstName
		}
		return u
	})
	if err != nil {
		return User{}, err
	}
	return newUser(m.deps, user), nil
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
