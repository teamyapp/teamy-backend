package resolver

import (
	"context"
	"github.com/teamyapp/teamy-backend/app/collect"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type Query struct {
	deps *Dependencies
}

func (q Query) Me(ct context.Context) (User, error) {
	panic("implement me")
}

func (q Query) Tasks(ct context.Context, args struct {
	Filter *TaskFilter
}) ([]Task, error) {
	panic("implement me")
}

func (q Query) Teams(ct context.Context, args struct {
	Filter *TeamFilter
}) ([]Team, error) {
	panic("implement me")
}

func (q Query) Invitations(ct context.Context, args struct {
	Filter *InvitationFilter
}) ([]Invitation, error) {
	invitations, err := q.deps.invitationDao.FindAllInvitations()
	if err != nil {
		return nil, err
	}

	if args.Filter != nil {
		invitations = collect.Filter(invitations, func(invitation entityv2.Invitation) bool {
			return matchInvitation(*args.Filter, invitation)
		})
	}

	return collect.Map(invitations, func(invitation entityv2.Invitation, _ int) Invitation {
		return newInvitation(q.deps, invitation)
	}), nil
}

func NewQuery(deps *Dependencies) Query {
	return Query{
		deps: deps,
	}
}
