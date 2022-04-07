package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type Team struct {
}

func (Team) ID(ctx context.Context) (graphql.ID, error) {
}

func (Team) Name(ctx context.Context) (string, error) {
}

func (Team) IconURL(ctx context.Context) (*string, error) {
}

func (Team) CreatedAt(ctx context.Context) (graphql.Time, error) {
}

func (Team) Creator(ctx context.Context) (User, error) {
}

func (Team) Owner(ctx context.Context) (User, error) {
}

func (Team) Members(ctx context.Context) ([]User, error) {
}

func (Team) Tasks(ctx context.Context, args struct {
	Filter TaskFilter
}) ([]Task, error) {
}

func (Team) Invitations(ctx context.Context) ([]Invitation, error) {
}
