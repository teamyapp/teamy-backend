package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/core/collect"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type User struct {
	deps *Dependencies
	user entity.User
}

func (u User) ID(ct context.Context) graphql.ID {
	return toGraphQLID(u.user.ID)
}

func (u User) FirstName(ct context.Context) string {
	return u.user.FirstName
}

func (u User) LastName(ct context.Context) string {
	return u.user.LastName
}

func (u User) ProfileURL(ct context.Context) *string {
	return u.user.ProfileURL
}

func (u User) CreatedAt(ct context.Context) graphql.Time {
	return toGraphQLTime(u.user.CreatedAt)
}

func (u User) Teams(ct context.Context, args struct {
	Filter *TeamFilter
}) ([]Team, error) {
	ids, err := u.deps.teamMemberDao.FindTeamIDsByUserID(u.user.ID)
	if err != nil {
		return nil, err
	}

	if len(ids) < 1 {
		return []Team{}, nil
	}

	teamEntities, err := u.deps.teamDao.FindTeamsByIDs(ids)
	if err != nil {
		return nil, err
	}

	if args.Filter != nil {
		teamEntities = collect.Filter(teamEntities, func(team entity.Team) bool {
			return matchTeam(*args.Filter, team)
		})
	}

	return collect.Map(teamEntities, func(team entity.Team, _ int) Team {
		return newTeam(u.deps, team)
	}), nil
}

func newUser(deps *Dependencies, user entity.User) User {
	return User{deps: deps, user: user}
}
