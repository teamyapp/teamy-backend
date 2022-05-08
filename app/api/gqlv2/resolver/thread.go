package resolver

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/collect"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type Thread struct {
	deps   Dependencies
	thread entityv2.Thread
}

func (t Thread) ID(ctx context.Context) graphql.ID {
	return toGraphQLID(t.thread.ID)
}

func (t Thread) Messages(ctx context.Context) ([]Message, error) {
	if t.thread.MessageIDs == nil {
		return nil, nil
	}

	messages, err := t.deps.messageDao.FindMessagesByIDs(t.thread.MessageIDs)
	if err != nil {
		return nil, err
	}

	return collect.Map(messages, func(messageEntity entityv2.Message, index int) Message {
		return Message{
			message: messageEntity,
			deps:    t.deps,
		}
	}), nil
}
