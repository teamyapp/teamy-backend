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

func (u User) Teams(ct context.Context) ([]Team, error) {
	ids, err := u.deps.teamMemberDao.FindTeamIDsByUserID(u.user.ID)
	if err != nil {
		return nil, err
	}

	teamEntities, err := u.deps.teamDao.FindTeamsByIDs(ids)
	if err != nil {
		return nil, err
	}

	teams := make([]Team, 0, 0)
	for _, teamEntity := range teamEntities {
		teams = append(teams, Team{
			team: teamEntity,
			deps: u.deps,
		})
	}

	return teams, nil
}
