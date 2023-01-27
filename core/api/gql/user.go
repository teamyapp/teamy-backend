package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/telemetry"
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

func (u User) UpdatedAt(ct context.Context) *graphql.Time {
	return toGraphQLTimePtr(u.user.UpdatedAt)
}

func (u User) Teams(ct context.Context, args struct {
	Filter *TeamFilter
}) ([]Team, error) {
	filter, err := fromGraphQLTeamFilterPtr(ct, u.deps.dataCollector, args.Filter)
	if err != nil {
		u.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	teams, err := u.deps.teamService.FindTeamsForUser(ct, u.user.ID, filter)
	if err != nil {
		u.deps.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
	}

	return collect.Map(teams, func(team entity.Team, _ int) Team {
		return newTeam(u.deps, team)
	}), nil
}

func newUser(deps *Dependencies, user entity.User) User {
	return User{deps: deps, user: user}
}
