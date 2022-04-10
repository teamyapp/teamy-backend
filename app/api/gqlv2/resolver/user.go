package resolver

import (
	"context"

	"github.com/teamyapp/teamy-backend/app/entityv2"

	"github.com/graph-gophers/graphql-go"
)

type User struct {
	deps Dependencies
	user entityv2.User
}

func (u User) ID(ctx context.Context) graphql.ID {
	return toGraphQLID(u.user.ID)
}

func (u User) FirstName(ctx context.Context) string {
	return u.user.FirstName
}

func (u User) LastName(ctx context.Context) string {
	return u.user.LastName
}

func (u User) ProfileURL(ctx context.Context) *string {
	return &u.user.ProfileURL
}

func (u User) CreatedAt(ctx context.Context) graphql.Time {
	return toGraphQLTime(u.user.CreatedAt)
}

func (u User) Teams(ctx context.Context) ([]Team, error) {
	ids, err := u.deps.teamMemberDao.FindTeamIDsForUser(u.user.ID)
	if err != nil {
		return nil, err
	}

	teams := make([]Team, 0, 0)
	for _, id := range ids {
		team, err := u.deps.teamDao.FindTeam(id)
		if err != nil {
			return nil, err
		}
		teams = append(teams, Team{
			team: team,
			deps: u.deps,
		})
	}

	return teams, nil
}
