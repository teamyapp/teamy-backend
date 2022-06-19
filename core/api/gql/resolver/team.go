package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Team struct {
	deps *Dependencies
	team entity.Team
}

func (t Team) ID(ct context.Context) graphql.ID {
	return toGraphQLID(t.team.ID)
}

func (t Team) Name(ct context.Context) string {
	return t.team.Name
}

func (t Team) IconURL(ct context.Context) *string {
	return t.team.IconURL
}

func (t Team) CreatedAt(ct context.Context) graphql.Time {
	return toGraphQLTime(t.team.CreatedAt)
}

func (t Team) Creator(ct context.Context) (User, error) {
	user, err := t.deps.userDao.FindUserByID(t.team.CreatorUserID)
	if err != nil {
		return User{}, nil
	}

	return newUser(t.deps, user), nil
}

func (t Team) Owner(ct context.Context) (User, error) {
	user, err := t.deps.userDao.FindUserByID(t.team.OwnerUserID)
	if err != nil {
		return User{}, nil
	}

	return newUser(t.deps, user), nil
}

func (t Team) Members(ct context.Context) ([]User, error) {
	teamMemberIDs, err := t.deps.teamMemberDao.FindTeamMemberIDsByTeamID(t.team.ID)
	if err != nil {
		return nil, err
	}

	userEntities, err := t.deps.userDao.FindUsersByIDs(teamMemberIDs)
	if err != nil {
		return nil, err
	}

	return collect.Map(userEntities, func(userEntity entity.User, _ int) User {
		return newUser(t.deps, userEntity)
	}), nil
}

func (t Team) Tasks(ct context.Context, args struct {
	Filter *TaskFilter
}) ([]Task, error) {
	tasks, err := t.deps.taskDao.FindTasksByTeamID(t.team.ID)
	if err != nil {
		return nil, err
	}

	if args.Filter != nil {
		tasks = collect.Filter(tasks, func(task entity.Task) bool {
			return matchTask(*args.Filter, task)
		})
	}

	return collect.Map(tasks, func(task entity.Task, _ int) Task {
		return newTask(t.deps, task)
	}), nil
}

func (t Team) Invitations(ct context.Context, args struct {
	Filter *InvitationFilter
}) ([]Invitation, error) {
	invitationEntities, err := t.deps.invitationDao.FindInvitationsByTeamID(t.team.ID)
	if err != nil {
		return nil, err
	}

	if args.Filter != nil {
		invitationEntities = collect.Filter(invitationEntities, func(invitationEntity entity.Invitation) bool {
			return matchInvitation(*args.Filter, invitationEntity)
		})
	}

	return collect.Map(invitationEntities, func(invitationEntity entity.Invitation, _ int) Invitation {
		return newInvitation(t.deps, invitationEntity)
	}), nil
}

func newTeam(deps *Dependencies, team entity.Team) Team {
	return Team{
		deps: deps,
		team: team,
	}
}
