package gql

import (
	"context"
	"log"

	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Query struct {
	deps *Dependencies
}

func (q Query) Me(ct context.Context) (User, error) {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	user, err := q.deps.userDao.FindUserByID(userID)
	if err != nil {
		log.Println(err)
		return User{}, err
	}

	return newUser(q.deps, user), err
}

func (q Query) Tasks(ct context.Context, args struct {
	Filter *TaskFilter
}) ([]Task, error) {
	tasks, err := q.deps.taskDao.FindAllTasks()
	if err != nil {
		return nil, err
	}

	if args.Filter != nil {
		tasks = collect.Filter(tasks, func(task entity.Task) bool {
			return matchTask(*args.Filter, task)
		})
	}

	return collect.Map(tasks, func(task entity.Task, _ int) Task {
		return newTask(q.deps, task)
	}), nil
}

func (q Query) Teams(ct context.Context, args struct {
	Filter *TeamFilter
}) ([]Team, error) {
	teams, err := q.deps.teamDao.FindAllTeams()
	if err != nil {
		return nil, err
	}

	if args.Filter != nil {
		teams = collect.Filter(teams, func(team entity.Team) bool {
			return matchTeam(*args.Filter, team)
		})
	}

	return collect.Map(teams, func(team entity.Team, _ int) Team {
		return newTeam(q.deps, team)
	}), nil
}

func (q Query) Invitations(ct context.Context, args struct {
	Filter *InvitationFilter
}) ([]Invitation, error) {
	invitations, err := q.deps.invitationDao.FindAllInvitations()
	if err != nil {
		return nil, err
	}

	if args.Filter != nil {
		invitations = collect.Filter(invitations, func(invitation entity.Invitation) bool {
			return matchInvitation(*args.Filter, invitation)
		})
	}

	return collect.Map(invitations, func(invitation entity.Invitation, _ int) Invitation {
		return newInvitation(q.deps, invitation)
	}), nil
}

func NewQuery(deps *Dependencies) Query {
	return Query{
		deps: deps,
	}
}
