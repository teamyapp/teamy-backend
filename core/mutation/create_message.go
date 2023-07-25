package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateMessage struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	messageDao       dao.Message
	taskDao          dao.Task
	id               uint64
	message          entity.Message
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*CreateMessage)(nil)

func (c *CreateMessage) GetID() uint64 {
	return c.id
}

func (c *CreateMessage) Execute(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return c.messageDao.CreateMessage(ct, tx, c.message)
}

func (c *CreateMessage) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifierPrepared {
		return nil
	}

	task, err := c.taskDao.FindTaskByCommentsThreadIDWithTx(ct, tx, c.message.ThreadID)
	if err != nil {
		return err
	}

	var internalErr *errs.Error
	c.clientNotifiers, internalErr = c.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
	if internalErr != nil {
		return internalErr
	}

	c.notifierPrepared = true
	return nil
}

func (c *CreateMessage) Undo() *errs.Error {
	return nil
}

func (c *CreateMessage) GetClientNotifiers() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateMessage) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.MessageCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.message,
	}
}

func (c *CreateMessage) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateMessage(
	stateSyncer *realtime.StateSyncer,
	messageDao dao.Message,
	taskDao dao.Task,
	logger telemetry.Logger,
	message entity.Message,
) *CreateMessage {
	return &CreateMessage{
		logger:           logger,
		stateSyncer:      stateSyncer,
		messageDao:       messageDao,
		taskDao:          taskDao,
		id:               stateSyncer.NextMutationID(),
		message:          message,
		notifierPrepared: false,
	}
}
