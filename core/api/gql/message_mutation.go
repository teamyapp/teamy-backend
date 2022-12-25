package gql

import (
	"context"
	"errors"
	"time"

	"github.com/graph-gophers/graphql-go"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/realtime/mutation"
)

func (m Mutation) CreateMessage(ct context.Context, args struct {
	ThreadID graphql.ID
	Message  struct {
		Body string
	}
}) (Message, error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	genMessageIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "messageID"}
	genMessageIDRes, err := m.deps.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genMessageIDReq)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	threadID, err := fromGraphQLID(args.ThreadID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	message := entity.Message{
		ID:           genMessageIDRes.UniqueNumber,
		Body:         args.Message.Body,
		ThreadID:     threadID,
		AuthorUserID: userID,
		CreatedAt:    time.Now(),
	}

	task, err := m.deps.taskDao.FindTaskByCommentsThreadID(ct, message.ThreadID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	transaction := realtime.NewTransaction(m.deps.stateSyncer, m.deps.dataCollector)
	createMessageMutation := mutation.NewCreateMessageMutation(
		task.OwningTeamID,
		m.deps.stateSyncer,
		message,
		m.deps.messageDao,
		m.deps.dataCollector)
	transaction.AddMutation(ct, createMessageMutation)
	err = transaction.Commit(ct)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	message, err := m.deps.messageDao.FindMessageByID(ct, messageID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	message.Body = args.Input.Body
	now := time.Now()
	message.UpdatedAt = &now

	task, err := m.deps.taskDao.FindTaskByCommentsThreadID(ct, message.ThreadID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	transaction := realtime.NewTransaction(m.deps.stateSyncer, m.deps.dataCollector)
	updateMessageMutation := mutation.NewUpdateMessageMutation(
		task.OwningTeamID,
		m.deps.stateSyncer,
		message,
		m.deps.messageDao,
		m.deps.dataCollector)
	transaction.AddMutation(ct, updateMessageMutation)
	err = transaction.Commit(ct)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	return newMessage(m.deps, message), nil
}

func (m Mutation) DeleteMessage(ct context.Context, args struct {
	MessageID graphql.ID
}) (Message, error) {
	messageID, err := fromGraphQLID(args.MessageID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	message, err := m.deps.messageDao.FindMessageByID(ct, messageID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	task, err := m.deps.taskDao.FindTaskByCommentsThreadID(ct, message.ThreadID)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}
	transaction := realtime.NewTransaction(m.deps.stateSyncer, m.deps.dataCollector)
	deleteMessageMutation := mutation.NewDeleteMessageMutation(
		task.OwningTeamID,
		m.deps.stateSyncer,
		message.ID,
		m.deps.messageDao,
		m.deps.dataCollector)
	transaction.AddMutation(ct, deleteMessageMutation)
	err = transaction.Commit(ct)
	if err != nil {
		m.deps.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return Message{}, err
	}

	return newMessage(m.deps, message), nil
}
