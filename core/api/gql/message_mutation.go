package gql

import (
	"context"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

func (m Mutation) CreateMessage(ct context.Context, args struct {
	ThreadID graphql.ID
	Message  struct {
		Body string
	}
}) (Message, error) {
	userID, err := ctx.UserIDFromContext(m.deps.dataCollector, ct)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	genMessageIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "messageID"}
	genMessageIDRes, err := m.deps.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genMessageIDReq)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	threadID, err := fromGraphQLID(args.ThreadID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	message := entity.Message{
		ID:           genMessageIDRes.UniqueNumber,
		Body:         args.Message.Body,
		ThreadID:     threadID,
		AuthorUserID: userID,
		CreatedAt:    time.Now(),
	}

	err = m.deps.messageSyncer.CreateAndSyncMessage(message)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	return newMessage(m.deps, message), nil
}

func (m Mutation) UpdateMessage(ct context.Context, args struct {
	MessageID graphql.ID
	Input     struct {
		Body string
	}
}) (Message, error) {
	messageID, err := fromGraphQLID(args.MessageID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	message, err := m.deps.messageDao.FindMessageByID(messageID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	message.Body = args.Input.Body
	now := time.Now()
	message.UpdatedAt = &now
	err = m.deps.messageSyncer.UpdateAndSyncMessage(message)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	return newMessage(m.deps, message), nil
}

func (m Mutation) DeleteMessage(ct context.Context, args struct {
	MessageID graphql.ID
}) (Message, error) {
	messageID, err := fromGraphQLID(args.MessageID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	message, err := m.deps.messageDao.FindMessageByID(messageID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	err = m.deps.messageSyncer.DeleteAndSyncMessage(messageID)
	if err != nil {
		m.deps.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	return newMessage(m.deps, message), nil
}
