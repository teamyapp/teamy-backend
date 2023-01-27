package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Client struct {
	deps   *Dependencies
	client entity.Client
}

func (c Client) ID() graphql.ID {
	return toGraphQLID(c.client.ID)
}

func (c Client) User(ct context.Context) (User, error) {
	user, err := c.deps.userService.FindUserByID(ct, c.client.UserID)
	if err != nil {
		c.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})

		return User{}, err
	}

	return newUser(c.deps, user), nil
}

func newClient(deps *Dependencies, client entity.Client) Client {
	return Client{
		deps:   deps,
		client: client,
	}
}
