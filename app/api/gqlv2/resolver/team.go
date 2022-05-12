package resolver

import (
	"context"
	"strings"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/collect"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type Team struct {
	deps Dependencies
	team entityv2.Team
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
	user, err := t.deps.userDao.FindUserByID(t.team.CreatorID)
	if err != nil {
		return User{}, nil
	}

	return User{
		user: user,
		deps: t.deps,
	}, nil
}

func (t Team) Owner(ct context.Context) (User, error) {
	user, err := t.deps.userDao.FindUserByID(t.team.OwnerID)
	if err != nil {
		return User{}, nil
	}

	return User{
		user: user,
		deps: t.deps,
	}, nil
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

	return collect.Map(userEntities, func(userEntity entityv2.User, _ int) User {
		return User{
			user: userEntity,
			deps: t.deps,
		}
	}), nil
}

func (t Team) Tasks(ct context.Context, args struct {
	Filter TaskFilter
}) ([]Task, error) {
	tasks, err := t.deps.taskDao.FindTasksByTeamID(t.team.ID)
	if err != nil {
		return nil, err
	}

	filteredTasks := collect.Filter(tasks, func(task entityv2.Task) bool {
		return matchTask(args.Filter, task)
	})

	return collect.Map(filteredTasks, func(filteredTask entityv2.Task, _ int) Task {
		return Task{
			task: filteredTask,
			deps: t.deps,
		}
	}), nil
}

func (t Team) Invitations(ct context.Context) ([]Invitation, error) {
	invitationEntities, err := t.deps.invitationDao.FindInvitationsByTeamID(t.team.ID)
	if err != nil {
		return nil, err
	}

	return collect.Map(invitationEntities, func(invitationEntity entityv2.Invitation, _ int) Invitation {
		return Invitation{
			invitation: invitationEntity,
			deps:       t.deps,
		}
	}), nil
}

func matchTask(filter TaskFilter, task entityv2.Task) bool {
	ownerID, err := fromGraphQLIDPtr(filter.OwnerID)
	if err != nil {
		return false
	}

	if filter.OwnerID != nil && ownerID != task.OwnerUserID {
		return false
	}

	if filter.Status != nil && *filter.Status != task.Status {
		return false
	}

	if filter.Goal != nil && strings.Contains(task.Goal, *filter.Goal) {
		return false
	}

	return true
}
