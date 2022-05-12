package resolver

import (
	"context"
)

type Query struct {
}

func (q Query) Me(ct context.Context) (User, error) {
	panic("implement me")
}

func (q Query) Tasks(ct context.Context, args struct {
	Filter TaskFilter
}) ([]Task, error) {
	panic("implement me")
}

func (q Query) Teams(ct context.Context, args struct {
	Filter TeamFilter
}) ([]Team, error) {
	panic("implement me")
}

func (q Query) Invitations(ct context.Context, args struct {
	Filter InvitationFilter
}) ([]Invitation, error) {
	panic("implement me")
}
