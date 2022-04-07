package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
)

type User struct {
}

func (User) ID(ctx context.Context) (graphql.ID, error) {
}

func (User) FirstName(ctx context.Context) (string, error) {
}

func (User) LastName(ctx context.Context) (string, error) {
}

func (User) ProfileURL(ctx context.Context) (*string, error) {
}

func (User) CreatedAt(ctx context.Context) (graphql.Time, error) {
}

func (User) Teams(ctx context.Context) ([]Team, error) {
}
