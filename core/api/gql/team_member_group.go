package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/service"
)

type TeamMemberGroup struct {
	deps        *Dependencies
	memberGroup entity.TeamMemberGroup
}

func (m TeamMemberGroup) ID() graphql.ID {
	return toGraphQLID(m.memberGroup.ID)
}

func (m TeamMemberGroup) Name() string {
	return m.memberGroup.Name
}

func (m TeamMemberGroup) Team(ct context.Context) Team {
	team, err := m.deps.teamService.FindTeamByID(ct, m.memberGroup.TeamID)
	if err != nil {
		m.deps.logger.Error(err)
		return Team{}
	}

	return newTeam(m.deps, team)
}

func (m TeamMemberGroup) Members(ct context.Context) ([]User, error) {
	users, err := m.deps.userService.FindUsersByIDs(ct, m.memberGroup.MemberUserIDs)
	if err != nil {
		m.deps.logger.Error(err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(users, func(user entity.User, _ int) User {
		return newUser(m.deps, user)
	}), nil
}

func (m TeamMemberGroup) CreatedAt() graphql.Time {
	return toGraphQLTime(m.memberGroup.CreatedAt)
}

func (m TeamMemberGroup) UpdatedAt() *graphql.Time {
	return toGraphQLTimePtr(m.memberGroup.UpdatedAt)
}

func newTeamMemberGroup(deps *Dependencies, memberGroup entity.TeamMemberGroup) TeamMemberGroup {
	return TeamMemberGroup{
		deps:        deps,
		memberGroup: memberGroup,
	}
}

func (m Mutation) CreateTeamMemberGroup(ct context.Context, args struct {
	TeamID graphql.ID
	Input  struct {
		Name string
	}
}) (TeamMemberGroup, error) {
	teamID, argErr := fromGraphQLID(args.TeamID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return TeamMemberGroup{}, errs.ToResolverErr(internalErr)
	}

	createTeamGroupInput := service.CreateTeamMemberGroupInput{
		TeamID: teamID,
		Name:   args.Input.Name,
	}

	teamMemberGroup, err := m.deps.teamService.CreateTeamMemberGroup(ct, createTeamGroupInput)
	if err != nil {
		m.deps.logger.Error(err)
		return TeamMemberGroup{}, errs.ToResolverErr(err)
	}

	return newTeamMemberGroup(m.deps, teamMemberGroup), nil
}

func (m Mutation) UpdateTeamMemberGroup(ct context.Context, args struct {
	ID    graphql.ID
	Input struct {
		Name string
	}
}) (TeamMemberGroup, error) {
	id, argErr := fromGraphQLID(args.ID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return TeamMemberGroup{}, errs.ToResolverErr(internalErr)
	}

	updateTeamGroupInput := service.UpdateTeamMemberGroupInput{
		GroupID: id,
		Name:    args.Input.Name,
	}

	teamMemberGroup, err := m.deps.teamService.UpdateTeamMemberGroup(ct, updateTeamGroupInput)
	if err != nil {
		m.deps.logger.Error(err)
		return TeamMemberGroup{}, errs.ToResolverErr(err)
	}

	return newTeamMemberGroup(m.deps, teamMemberGroup), nil
}

func (m Mutation) DeleteTeamMemberGroup(ct context.Context, args struct {
	ID graphql.ID
}) (TeamMemberGroup, error) {
	id, argErr := fromGraphQLID(args.ID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return TeamMemberGroup{}, errs.ToResolverErr(internalErr)
	}

	teamMemberGroup, err := m.deps.teamService.DeleteTeamMemberGroup(ct, id)
	if err != nil {
		m.deps.logger.Error(err)
		return TeamMemberGroup{}, errs.ToResolverErr(err)
	}

	return newTeamMemberGroup(m.deps, teamMemberGroup), nil
}

func (m Mutation) AddUserToTeamMemberGroup(ct context.Context, args struct {
	ID     graphql.ID
	UserID graphql.ID
}) (User, error) {
	id, argErr := fromGraphQLID(args.ID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return User{}, errs.ToResolverErr(internalErr)
	}

	userID, argErr := fromGraphQLID(args.UserID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return User{}, errs.ToResolverErr(internalErr)
	}

	user, err := m.deps.teamService.AddUserToTeamMemberGroup(ct, id, userID)
	if err != nil {
		m.deps.logger.Error(err)
		return User{}, errs.ToResolverErr(err)
	}

	return newUser(m.deps, user), nil
}

func (m Mutation) RemoveUserFromTeamMemberGroup(ct context.Context, args struct {
	ID     graphql.ID
	UserID graphql.ID
}) (User, error) {
	id, argErr := fromGraphQLID(args.ID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return User{}, errs.ToResolverErr(internalErr)
	}

	userID, argErr := fromGraphQLID(args.UserID)
	if argErr != nil {
		internalErr := errs.NewError(
			errs.InvalidArgument,
			argErr.Error(),
		)
		m.deps.logger.ErrorWithContext(ct, internalErr)
		return User{}, errs.ToResolverErr(internalErr)
	}

	user, err := m.deps.teamService.RemoveUserFromTeamMemberGroup(ct, id, userID)
	if err != nil {
		m.deps.logger.Error(err)
		return User{}, errs.ToResolverErr(err)
	}

	return newUser(m.deps, user), nil
}
