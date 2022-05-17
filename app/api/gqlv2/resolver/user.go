package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type User struct {
	deps *Dependencies
	user entityv2.User
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

	if args.Filter.TeamID != nil {
		teamID, err := fromGraphQLIDPtr(args.Filter.TeamID)
		if err != nil {
			return nil, err
		}

		teamEntity, err := u.deps.teamDao.FindTeamByID(*teamID)
		if err != nil {
			return nil, err
		}
		return []Team{newTeam(u.deps, teamEntity)}, nil
	}

	teamEntities, err := u.deps.teamDao.FindTeamsByIDs(ids)
	if err != nil {
		return nil, err
	}

	teams := make([]Team, 0, 0)
	for _, teamEntity := range teamEntities {
		teams = append(teams, newTeam(u.deps, teamEntity))
	}

	return teams, nil
}

func newUser(deps *Dependencies, user entityv2.User) User {
	return User{deps: deps, user: user}
}
