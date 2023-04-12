package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/api/gql/scalar"
	"github.com/teamyapp/teamy-backend/core/service"
)

func (m Mutation) AddMemberToTeam(ct context.Context, args struct {
	TeamID       graphql.ID
	MemberUserID graphql.ID
}) (TeamMember, error) {
	teamID, argErr := fromGraphQLID(args.TeamID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return TeamMember{}, errs.ToResolverErr(internalErr)
	}

	memberUserID, argErr := fromGraphQLID(args.MemberUserID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return TeamMember{}, errs.ToResolverErr(internalErr)
	}

	teamMember, err := m.deps.teamService.AddMemberToTeam(ct, teamID, memberUserID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return TeamMember{}, errs.ToResolverErr(err)
	}

	return newTeamMember(m.deps, teamMember), nil
}

func (m Mutation) UpdateTeamMember(ct context.Context, args struct {
	TeamID graphql.ID
	Input  struct {
		UserID          graphql.ID
		WeeklyBandwidth scalar.Duration
	}
}) (TeamMember, error) {
	teamID, argErr := fromGraphQLID(args.TeamID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return TeamMember{}, errs.ToResolverErr(internalErr)
	}

	memberUserID, argErr := fromGraphQLID(args.Input.UserID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return TeamMember{}, errs.ToResolverErr(internalErr)
	}

	input := service.UpdateTeamMemberInput{
		UserID:          memberUserID,
		WeeklyBandwidth: args.Input.WeeklyBandwidth.Duration,
	}
	teamMember, err := m.deps.teamService.UpdateTeamMember(ct, teamID, input)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return TeamMember{}, errs.ToResolverErr(err)
	}

	return newTeamMember(m.deps, teamMember), nil
}

func (m Mutation) RemoveMemberFromTeam(ct context.Context, args struct {
	TeamID       graphql.ID
	MemberUserID graphql.ID
}) (TeamMember, error) {
	teamID, argErr := fromGraphQLID(args.TeamID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return TeamMember{}, errs.ToResolverErr(internalErr)
	}

	memberUserID, argErr := fromGraphQLID(args.MemberUserID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return TeamMember{}, errs.ToResolverErr(internalErr)
	}

	teamMember, err := m.deps.teamService.RemoveMemberFromTeam(ct, teamID, memberUserID)
	if err != nil {
		m.deps.logger.ErrorWithContext(ct, err)
		return TeamMember{}, errs.ToResolverErr(err)
	}

	return newTeamMember(m.deps, teamMember), nil
}
