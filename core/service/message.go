package service

import (
	"context"
	"errors"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateMessageInput struct {
	Body string
}

type UpdateMessageInput struct {
	Body string
}

type Message struct {
	dataCollector       obs.DataCollector
	cloudClientRegistry *cloudAPI.ClientRegistry
	stateSyncer         *realtime.StateSyncer
	messageDao          dao.Message
	taskDao             dao.Task
}

func (m Message) FindMessages(ct context.Context, threadID uint64) ([]entity.Message, error) {
	return m.messageDao.FindMessagesByThreadID(ct, threadID)
}

func (m Message) CreateMessage(ct context.Context, threadID uint64, input CreateMessageInput) (entity.Message, error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Message{}, err
	}

	genMessageIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "messageID"}
	genMessageIDRes, err := m.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genMessageIDReq)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Message{}, err
	}

	message := entity.Message{
		ID:           genMessageIDRes.UniqueNumber,
		Body:         input.Body,
		ThreadID:     threadID,
		AuthorUserID: userID,
		CreatedAt:    time.Now(),
	}
	realTimeTransaction := realtime.NewTransaction(m.dataCollector, m.stateSyncer)
	createMessageMutation := mutation.NewCreateMessageMutation(
		m.stateSyncer,
		m.messageDao,
		m.taskDao,
		m.dataCollector,
		message)
	err = realTimeTransaction.ApplyMutation(ct, createMessageMutation)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Message{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Message{}, err
	}

	return message, nil
}

func (m Message) UpdateMessage(ct context.Context, messageID uint64, input UpdateMessageInput) (entity.Message, error) {
	message, err := m.messageDao.FindMessageByID(ct, messageID)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Message{}, err
	}

	message.Body = input.Body
	now := time.Now()
	message.UpdatedAt = &now
	realTimeTransaction := realtime.NewTransaction(m.dataCollector, m.stateSyncer)
	updateMessageMutation := mutation.NewUpdateMessageMutation(
		m.dataCollector,
		m.stateSyncer,
		m.messageDao,
		m.taskDao,
		message)
	err = realTimeTransaction.ApplyMutation(ct, updateMessageMutation)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Message{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Message{}, err
	}

	return message, nil
}

func (m Message) DeleteMessage(ct context.Context, messageID uint64) (entity.Message, error) {
	message, err := m.messageDao.FindMessageByID(ct, messageID)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Message{}, err
	}
	realTimeTransaction := realtime.NewTransaction(m.dataCollector, m.stateSyncer)
	deleteMessageMutation := mutation.NewDeleteMessageMutation(
		m.dataCollector,
		m.stateSyncer,
		m.messageDao,
		m.taskDao,
		message)
	err = realTimeTransaction.ApplyMutation(ct, deleteMessageMutation)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Message{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		m.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Message{}, err
	}

	return message, nil
}

func NewMessage(
	dataCollector obs.DataCollector,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	stateSyncer *realtime.StateSyncer,
	messageDao dao.Message,
	taskDao dao.Task,
) Message {
	return Message{
		dataCollector:       dataCollector,
		cloudClientRegistry: cloudClientRegistry,
		stateSyncer:         stateSyncer,
		messageDao:          messageDao,
		taskDao:             taskDao,
	}
}
