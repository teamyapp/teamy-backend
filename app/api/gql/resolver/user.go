package resolver

import (
	"context"
	"log"

	"github.com/graph-gophers/graphql-go"
	"github.com/pkg/errors"
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/one/identity"
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
		userID := u.ID()
		args.Input.CreatorID = &userID
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

//////////////
// Mutation //
//////////////
// Admin && Debug Only
func (m Mutation) CreateUser(
	args struct {
		Input struct {
			ID         graphql.ID
			FirstName  *string
			LastName   *string
			ProfileURL *string
		}
	},
) (UserUpdate, error) {
	id, err := fromGraphQLID(args.Input.ID)
	if err != nil {
		return UserUpdate{}, err
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
	if args.Input.ProfileURL != nil {
		profileURL = *args.Input.ProfileURL
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
		return UserUpdate{}, err
	}
	return UserUpdate{deps: m.deps, user: user}, nil
}

func (m Mutation) User(ctx context.Context, args struct{ ID graphql.ID }) (UserUpdate, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return UserUpdate{}, err
	}

	entityID, err := fromGraphQLID(args.ID)
	if err != nil {
		return UserUpdate{}, err
	}

	if userID != entityID {
		return UserUpdate{},
			errors.Errorf("the logged in user %v is not allowed to mutate user %v", userID, entityID)
	}

	user, err := m.deps.Data.GetUser(entityID)
	if err != nil {
		return UserUpdate{}, err
	}

	return UserUpdate{
		deps: m.deps,
		user: user,
	}, nil
}

type UserUpdate struct {
	deps *Dependencies
	user entity.User
}

func (up UserUpdate) User() User {
	return newUser(up.deps, up.user)
}

func (up UserUpdate) UpdateActiveTeam(ctx context.Context, args struct {
	TeamID graphql.ID
}) (UserUpdate, error) {
	user, err := up.deps.Data.GetUser(up.user.ID)
	if err != nil {
		return UserUpdate{}, err
	}
	id, err := fromGraphQLID(args.TeamID)
	if err != nil {
		return UserUpdate{}, err
	}

	teams := up.deps.Data.FilterTeams(func(t entity.Team) bool { return t.ID == id })
	if len(teams) == 0 {
		return UserUpdate{}, errors.Errorf("team %v does not exist", id)
	}

	user, err = up.deps.Data.UpdateUser((user.ID), func(u entity.User) entity.User {
		u.ActiveTeamID = id
		return u
	})
	if err != nil {
		return UserUpdate{}, err
	}

	up.user = user
	return up, err
}

func (up UserUpdate) UpdateUser(
	ctx context.Context,
	args struct {
		Input struct {
			FirstName  *string
			LastName   *string
			ProfileUrl *string
		}
	},
) (UserUpdate, error) {
	user, err := up.deps.Data.UpdateUser(up.user.ID, func(u entity.User) entity.User {
		if args.Input.LastName != nil {
			u.LastName = *args.Input.LastName
		}
		if args.Input.FirstName != nil {
			u.FirstName = *args.Input.FirstName
		}
		if args.Input.ProfileUrl != nil {
			u.ProfileURL = *args.Input.ProfileUrl
		}
		return u
	})
	if err != nil {
		return UserUpdate{}, err
	}
	up.user = user
	return up, nil
}

//
// Deprecated
//

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

func (m Mutation) UpdateActiveTeam(ctx context.Context, args struct {
	TeamID graphql.ID
}) (User, error) {
	userID, err := identity.FromContext(ctx)
	if err != nil {
		return User{}, err
	}

	user, err := m.deps.Data.GetUser(userID)
	if err != nil {
		return User{}, err
	}
	id, err := fromGraphQLID(args.TeamID)
	if err != nil {
		return User{}, err
	}

	teams := m.deps.Data.FilterTeams(func(t entity.Team) bool { return t.ID == id })
	if len(teams) == 0 {
		return User{}, errors.Errorf("team %v does not exist", id)
	}

	user, err = m.deps.Data.UpdateUser((user.ID), func(u entity.User) entity.User {
		u.ActiveTeamID = id
		return u
	})
	if err != nil {
		log.Printf("%+v\n", err)
		return User{}, err
	}
	return newUser(m.deps, user), err
}
