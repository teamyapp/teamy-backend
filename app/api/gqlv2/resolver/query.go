package resolver

import (
	"context"
)

type Query struct {
}

func (q Query) Me(ctx context.Context) (User, error) {
}

func (q Query) Tasks(ctx context.Context, args struct {
	Filter TaskFilter
}) ([]Task, error) {
}

func (q Query) Teams(ctx context.Context, args struct {
	Filter TeamFilter
}) ([]Team, error) {
}

func (q Query) Invitations(ctx context.Context, args struct {
	Filter InvitationFilter
}) ([]Invitation, error) {
}
