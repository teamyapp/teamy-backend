package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/api/gql/scalar"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamMember struct {
	deps       *Dependencies
	teamMember entity.TeamMember
}

func (t TeamMember) Team(ct context.Context) (Team, error) {
	team, err := t.deps.teamService.FindTeamByID(ct, t.teamMember.TeamID)
	if err != nil {
		t.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return Team{}, errs.ToResolverErr(err)
	}

	return newTeam(t.deps, team), nil
}

func (t TeamMember) User(ct context.Context) (User, error) {
	user, err := t.deps.userService.FindUserByID(ct, t.teamMember.UserID)
	if err != nil {
		t.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return User{}, errs.ToResolverErr(err)
	}

	return newUser(t.deps, user), nil
}

func (t TeamMember) WeeklyBandwidth() scalar.Duration {
	return toGraphQLDuration(t.teamMember.WeeklyBandwidth)
}

func (t TeamMember) CreatedAt() graphql.Time {
	return toGraphQLTime(t.teamMember.CreatedAt)
}

func (t TeamMember) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(t.teamMember.UpdatedAt)
}

func newTeamMember(
	deps *Dependencies,
	teamMember entity.TeamMember,
) TeamMember {
	return TeamMember{
		deps:       deps,
		teamMember: teamMember,
	}
}
