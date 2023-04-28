package service

import (
	"context"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
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
	logger              telemetry.Logger
	cloudClientRegistry *cloudAPI.ClientRegistry
	stateSyncer         *realtime.StateSyncer
	transactionFactory  transaction.Factory
	toggles             feature.Toggles
	taskDaoV2           daov2.Task
	threadDaoV2         daov2.Thread
	messageDaoV2        daov2.Message
}

func (t Thread) CreateThread(ct context.Context) (uint64, *errs.Error) {
	genThreadIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "threadID"}
	genThreadIDRes, rpcErr := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genThreadIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return 0, internalErr
	}

	threadID := genThreadIDRes.UniqueNumber
	txCtx := TransactionsContext{
		logger:             t.logger,
		transactionFactory: t.transactionFactory,
		stateSyncer:        t.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return t.threadDaoV2.CreateThread(ct, tx, threadID)
	})
	return threadID, err
}

func (t Thread) FindMessages(ct context.Context, threadID uint64) ([]entity.Message, *errs.Error) {
	return t.messageDaoV2.FindMessagesByThreadID(ct, threadID)
}

func (t Thread) CreateMessage(ct context.Context, threadID uint64, input CreateMessageInput) (entity.Message, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Message{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	genMessageIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "messageID"}
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
	txCtx := TransactionsContext{
		logger:             t.logger,
		transactionFactory: t.transactionFactory,
		stateSyncer:        t.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		createMessageMutation := mutation.NewCreateMessage(
			t.stateSyncer,
			t.messageDaoV2,
			t.taskDaoV2,
			t.logger,
			message)
		err := createMessageMutation.ExecuteV2(ct, tx)
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
	message, err := t.messageDaoV2.FindMessageByID(ct, messageID)
	if err != nil {
		return entity.Message{}, err
	}

	message.Body = input.Body
	now := time.Now().UTC()
	message.UpdatedAt = &now
	txCtx := TransactionsContext{
		logger:             t.logger,
		transactionFactory: t.transactionFactory,
		stateSyncer:        t.stateSyncer,
		ct:                 ct,
	}
	err = txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		updateMessageMutation := mutation.NewUpdateMessage(
			t.logger,
			t.stateSyncer,
			t.messageDaoV2,
			t.taskDaoV2,
			message)
		err = updateMessageMutation.ExecuteV2(ct, tx)
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
	message, err := t.messageDaoV2.FindMessageByID(ct, messageID)
	if err != nil {
		return entity.Message{}, err
	}

	txCtx := TransactionsContext{
		logger:             t.logger,
		transactionFactory: t.transactionFactory,
		stateSyncer:        t.stateSyncer,
		ct:                 ct,
	}
	err = txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		deleteMessageMutation := mutation.NewDeleteMessage(
			t.logger,
			t.stateSyncer,
			t.messageDaoV2,
			t.taskDaoV2,
			message)
		err = deleteMessageMutation.ExecuteV2(ct, tx)
		if err != nil {
			return err
		}

		rtTx.AppendMutation(deleteMessageMutation)
		return nil
	})

	if err != nil {
		return entity.Message{}, err
	}

	return message, nil
}

func NewThread(
	logger telemetry.Logger,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	stateSyncer *realtime.StateSyncer,
	transactionFactory transaction.Factory,
	toggles feature.Toggles,
	taskDaoV2 daov2.Task,
	threadDaoV2 daov2.Thread,
	messageDaoV2 daov2.Message,
) Thread {
	return Thread{
		logger:              logger,
		cloudClientRegistry: cloudClientRegistry,
		stateSyncer:         stateSyncer,
		transactionFactory:  transactionFactory,
		toggles:             toggles,
		taskDaoV2:           taskDaoV2,
		threadDaoV2:         threadDaoV2,
		messageDaoV2:        messageDaoV2,
	}
}
