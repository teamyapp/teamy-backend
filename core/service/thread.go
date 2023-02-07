package service

import (
	"context"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
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

type Thread struct {
	dataCollector       telemetry.DataCollector
	cloudClientRegistry *cloudAPI.ClientRegistry
	stateSyncer         *realtime.StateSyncer
	taskDao             dao.Task
	threadDao           dao.Thread
	messageDao          dao.Message
}

func (t Thread) CreateThread(ct context.Context) (uint64, *errs.Error) {
	genThreadIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "threadID"}
	genThreadIDRes, rpcErr := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genThreadIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return 0, internalErr
	}

	threadID := genThreadIDRes.UniqueNumber
	err := t.threadDao.CreateThread(ct, threadID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return threadID, err
}

func (t Thread) FindMessages(ct context.Context, threadID uint64) ([]entity.Message, *errs.Error) {
	return t.messageDao.FindMessagesByThreadID(ct, threadID)
}

func (t Thread) CreateMessage(ct context.Context, threadID uint64, input CreateMessageInput) (entity.Message, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: "user ID not found",
		}
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.Message{}, internalErr
	}

	genMessageIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "messageID"}
	genMessageIDRes, rpcErr := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genMessageIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.Message{}, internalErr
	}

	message := entity.Message{
		ID:           genMessageIDRes.UniqueNumber,
		Body:         input.Body,
		ThreadID:     threadID,
		AuthorUserID: userID,
		CreatedAt:    time.Now(),
	}
	realTimeTransaction := realtime.NewTransaction(t.dataCollector, t.stateSyncer)
	createMessageMutation := mutation.NewCreateMessageMutation(
		t.stateSyncer,
		t.messageDao,
		t.taskDao,
		t.dataCollector,
		message)
	err := realTimeTransaction.ApplyMutation(ct, createMessageMutation)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.Message{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.Message{}, err
	}

	return message, nil
}

func (t Thread) UpdateMessage(ct context.Context, messageID uint64, input UpdateMessageInput) (entity.Message, *errs.Error) {
	message, err := t.messageDao.FindMessageByID(ct, messageID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.Message{}, err
	}

	message.Body = input.Body
	now := time.Now()
	message.UpdatedAt = &now
	realTimeTransaction := realtime.NewTransaction(t.dataCollector, t.stateSyncer)
	updateMessageMutation := mutation.NewUpdateMessageMutation(
		t.dataCollector,
		t.stateSyncer,
		t.messageDao,
		t.taskDao,
		message)
	err = realTimeTransaction.ApplyMutation(ct, updateMessageMutation)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.Message{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.Message{}, err
	}

	return message, nil
}

func (t Thread) DeleteMessage(ct context.Context, messageID uint64) (entity.Message, *errs.Error) {
	message, err := t.messageDao.FindMessageByID(ct, messageID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.Message{}, err
	}
	realTimeTransaction := realtime.NewTransaction(t.dataCollector, t.stateSyncer)
	deleteMessageMutation := mutation.NewDeleteMessageMutation(
		t.dataCollector,
		t.stateSyncer,
		t.messageDao,
		t.taskDao,
		message)
	err = realTimeTransaction.ApplyMutation(ct, deleteMessageMutation)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.Message{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return entity.Message{}, err
	}

	return message, nil
}

func NewThread(
	dataCollector telemetry.DataCollector,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	stateSyncer *realtime.StateSyncer,
	taskDao dao.Task,
	threadDao dao.Thread,
	messageDao dao.Message,
) Thread {
	return Thread{
		dataCollector:       dataCollector,
		cloudClientRegistry: cloudClientRegistry,
		stateSyncer:         stateSyncer,
		taskDao:             taskDao,
		threadDao:           threadDao,
		messageDao:          messageDao,
	}
}
