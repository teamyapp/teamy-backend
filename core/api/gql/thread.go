package gql

import (
	"context"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Thread struct {
	deps     *Dependencies
	threadID uint64
}

func (t Thread) ID() graphql.ID {
	return toGraphQLID(t.threadID)
}

func (t Thread) Messages(ct context.Context) ([]Message, error) {
	messages, err := t.deps.threadService.FindMessages(ct, t.threadID)
	if err != nil {
		t.deps.logger.ErrorWithContext(ct, err)
		return nil, errs.ToResolverErr(err)
	}

	return collect.Map(messages, func(message entity.Message, _ int) Message {
		return newMessage(t.deps, message)
	}), nil
}

func newThread(deps *Dependencies, threadID uint64) Thread {
	return Thread{
		deps:     deps,
		threadID: threadID,
	}
}
