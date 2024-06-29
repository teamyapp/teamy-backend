package service

import (
	"context"
	"time"

	"github.com/teamyapp/cloud/app/client"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	pbcloud "github.com/teamyapp/protocol/pb/pbgo/cloud"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/transaction"
)

type CreateMessageInput struct {
	Body string
}

type UpdateMessageInput struct {
	Body *string
}

type Thread struct {
	logger                  telemetry.Logger
	transactionGroupFactory transaction.GroupFactory
	toggles                 feature.Toggles
	cloudClientRegistry     *client.Registry
	stateSyncer             *realtime.StateSyncer
	transactionFactory      cloudTransaction.Factory
	taskDao                 dao.Task
	threadDao               dao.Thread
	messageDao              dao.Message
}

func (t Thread) CreateThread(ct context.Context) (uint64, *errs.Error) {
	// TODO: add authorization logic
	genThreadIDReq := &pbcloud.GenerateUniqueNumberRequest{SequenceName: "threadID"}
	genThreadIDRes, rpcErr := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genThreadIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return 0, internalErr
	}

	threadID := genThreadIDRes.UniqueNumber
	err := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			return t.threadDao.CreateThread(ct, tx, threadID)
		})
	return threadID, err
}

func (t Thread) FindMessages(ct context.Context, threadID uint64) ([]entity.Message, *errs.Error) {
	// TODO: add authorization logic
	return t.messageDao.FindMessagesByThreadID(ct, threadID)
}

func (t Thread) CreateMessage(ct context.Context, threadID uint64, input CreateMessageInput) (entity.Message, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Message{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	// TODO: add authorization logic
	genMessageIDReq := &pbcloud.GenerateUniqueNumberRequest{SequenceName: "messageID"}
	genMessageIDRes, rpcErr := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genMessageIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.Message{}, internalErr
	}

	message := entity.Message{
		ID:           genMessageIDRes.UniqueNumber,
		Body:         input.Body,
		ThreadID:     threadID,
		AuthorUserID: userID,
		CreatedAt:    time.Now(),
	}
	err := t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			createMessageMutation := mutation.NewCreateMessage(
				t.stateSyncer,
				t.messageDao,
				t.taskDao,
				t.logger,
				message)
			err := createMessageMutation.Execute(ct, tx)
			if err != nil {
				return err
			}

			rtTx.AppendMutation(createMessageMutation)
			return nil
		})

	if err != nil {
		return entity.Message{}, err
	}

	return message, nil
}

func (t Thread) UpdateMessage(ct context.Context, messageID uint64, input UpdateMessageInput) (entity.Message, *errs.Error) {
	// TODO: add authorization logic
	message, err := t.messageDao.FindMessageByID(ct, messageID)
	if err != nil {
		return entity.Message{}, err
	}

	if input.Body != nil {
		message.Body = *input.Body
	}
	now := time.Now().UTC()
	message.UpdatedAt = &now
	err = t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			updateMessageMutation := mutation.NewUpdateMessage(
				t.logger,
				t.stateSyncer,
				t.messageDao,
				t.taskDao,
				message)
			err = updateMessageMutation.Execute(ct, tx)
			if err != nil {
				return err
			}

			rtTx.AppendMutation(updateMessageMutation)
			return nil
		})

	if err != nil {
		return entity.Message{}, err
	}

	return message, nil
}

func (t Thread) DeleteMessage(ct context.Context, messageID uint64) (entity.Message, *errs.Error) {
	// TODO: add authorization logic
	message, err := t.messageDao.FindMessageByID(ct, messageID)
	if err != nil {
		return entity.Message{}, err
	}

	err = t.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			deleteMessageMutation := mutation.NewDeleteMessage(
				t.logger,
				t.stateSyncer,
				t.messageDao,
				t.taskDao,
				message)
			err = deleteMessageMutation.Execute(ct, tx)
			if err != nil {
				return err
			}

			rtTx.AppendMutation(deleteMessageMutation)
			return nil
		})

	if err != nil {
		return entity.Message{}, err
	}

	// TODO: clean up resource relations in authorization service
	return message, nil
}

func NewThread(
	logger telemetry.Logger,
	transactionGroupFactory transaction.GroupFactory,
	toggles feature.Toggles,
	cloudClientRegistry *client.Registry,
	stateSyncer *realtime.StateSyncer,
	transactionFactory cloudTransaction.Factory,
	taskDao dao.Task,
	threadDao dao.Thread,
	messageDao dao.Message,
) Thread {
	return Thread{
		logger:                  logger,
		transactionGroupFactory: transactionGroupFactory,
		toggles:                 toggles,
		cloudClientRegistry:     cloudClientRegistry,
		stateSyncer:             stateSyncer,
		transactionFactory:      transactionFactory,
		taskDao:                 taskDao,
		threadDao:               threadDao,
		messageDao:              messageDao,
	}
}
