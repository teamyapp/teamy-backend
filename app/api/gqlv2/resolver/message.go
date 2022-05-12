package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type Message struct {
	deps    Dependencies
	message entityv2.Message
}

func (m Message) ID(ct context.Context) graphql.ID {
	return toGraphQLID(m.message.ID)
}

func (m Message) Body(ct context.Context) string {
	return m.message.Body
}

func (m Message) CreatedAt(ct context.Context) graphql.Time {
	return toGraphQLTime(m.message.CreatedAt)
}

func (m Message) UpdatedAt(ct context.Context) *graphql.Time {
	return toGraphQLTimePtr(m.message.UpdatedAt)
}
