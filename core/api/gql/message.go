package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Message struct {
	deps    *Dependencies
	message entity.Message
}

func (m Message) ID(ct context.Context) graphql.ID {
	return toGraphQLID(m.message.ID)
}

func (m Message) Body(ct context.Context) string {
	return m.message.Body
}

func (m Message) Author(ct context.Context) (User, error) {
	user, err := m.deps.userService.FindUserByID(ct, m.message.AuthorUserID)
	if err != nil {
		m.deps.dataCollector.Logger.ErrorWithContext(ct, err)
		return User{}, errs.ToResolverErr(err)
	}

	return newUser(m.deps, user), nil
}

func (m Message) Thread(ct context.Context) Thread {
	return newThread(m.deps, m.message.ThreadID)
}

func (m Message) CreatedAt(ct context.Context) graphql.Time {
	return toGraphQLTime(m.message.CreatedAt)
}

func (m Message) UpdatedAt(ct context.Context) *graphql.Time {
	return toGraphQLTimePtr(m.message.UpdatedAt)
}

func newMessage(deps *Dependencies, message entity.Message) Message {
	return Message{
		deps:    deps,
		message: message,
	}
}
