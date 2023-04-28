package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateMessage struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	messageDaoV2     daov2.Message
	taskDaoV2        daov2.Task
	id               uint64
	message          entity.Message
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*CreateMessage)(nil)

func (c *CreateMessage) GetID() uint64 {
	return c.id
}

func (c *CreateMessage) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return c.messageDaoV2.CreateMessage(ct, tx, c.message)
}

func (c *CreateMessage) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
	if c.notifierPrepared {
		return nil
	}

	task, err := c.taskDaoV2.FindTaskByCommentsThreadIDWithTx(ct, tx, c.message.ThreadID)
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

func (c *CreateMessage) GetClientNotifiersV2() []*realtime.ClientNotifier {
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
	messageDaoV2 daov2.Message,
	taskDaoV2 daov2.Task,
	logger telemetry.Logger,
	message entity.Message,
) *CreateMessage {
	return &CreateMessage{
		logger:           logger,
		stateSyncer:      stateSyncer,
		messageDaoV2:     messageDaoV2,
		taskDaoV2:        taskDaoV2,
		id:               stateSyncer.NextMutationID(),
		message:          message,
		notifierPrepared: false,
	}
}
