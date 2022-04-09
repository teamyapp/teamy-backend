package resolver

import (
	"context"

	"github.com/teamyapp/teamy-backend/app/dao"

	"github.com/graph-gophers/graphql-go"
)

type User struct {
	user dao.User
	deps *Dependencies
}

func (User) ID(ctx context.Context) (graphql.ID, error) {
	panic("implement me")
}

func (User) FirstName(ctx context.Context) (string, error) {
	panic("implement me")
}

func (User) LastName(ctx context.Context) (string, error) {
	panic("implement me")
}

func (User) ProfileURL(ctx context.Context) (*string, error) {
	panic("implement me")
}

func (User) CreatedAt(ctx context.Context) (graphql.Time, error) {
	panic("implement me")
}

func (User) Teams(ctx context.Context) ([]Team, error) {
	panic("implement me")
}
