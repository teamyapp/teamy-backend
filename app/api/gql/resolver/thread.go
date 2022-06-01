package resolver

import (
	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/teamy-backend/app/collect"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type Thread struct {
	deps     *Dependencies
	threadID uint64
}

func (t Thread) ID() graphql.ID {
	return toGraphQLID(t.threadID)
}

func (t Thread) Messages() ([]Message, error) {
	messages, err := t.deps.messageDao.FindMessagesByThreadID(t.threadID)
	if err != nil {
		return nil, err
	}

	return collect.Map(messages, func(message entityv2.Message, _ int) Message {
		return newMessage(t.deps, message)
	}), nil
}

func newThread(deps *Dependencies, threadID uint64) Thread {
	return Thread{
		deps:     deps,
		threadID: threadID,
	}
}
