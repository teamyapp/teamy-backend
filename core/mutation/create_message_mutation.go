package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateMessageMutation struct {
	logger           telemetry.Logger
	stateSyncer      *realtime.StateSyncer
	messageDao       dao.Message
	messageDaoV2     daov2.Message
	taskDao          dao.Task
	taskDaoV2        daov2.Task
	id               uint64
	message          entity.Message
	clientNotifiers  []*realtime.ClientNotifier
	notifierPrepared bool
}

var _ realtime.Mutation = (*CreateMessageMutation)(nil)

func (c *CreateMessageMutation) GetID() uint64 {
	return c.id
}

func (c *CreateMessageMutation) ExecuteV2(ct context.Context, tx *transaction.Transaction) *errs.Error {
	return c.messageDaoV2.CreateMessage(ct, tx, c.message)
}

func (c *CreateMessageMutation) PrepareClientNotifiers(ct context.Context, tx *transaction.Transaction) *errs.Error {
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

func (c *CreateMessageMutation) Execute(ct context.Context) *errs.Error {
	err := c.messageDao.CreateMessage(ct, c.message)
	if err != nil {
		c.logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (c *CreateMessageMutation) Undo() *errs.Error {
	return nil
}

func (c *CreateMessageMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	task, err := c.taskDao.FindTaskByCommentsThreadID(ct, c.message.ThreadID)
	if err != nil {
		c.logger.ErrorWithContext(ct, err)
		return []*realtime.ClientNotifier{}, err
	}

	return c.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
}

func (c *CreateMessageMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	return c.clientNotifiers
}

func (c *CreateMessageMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.MessageCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.message,
	}
}

func (c *CreateMessageMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateMessageMutation(
	stateSyncer *realtime.StateSyncer,
	messageDao dao.Message,
	messageDaoV2 daov2.Message,
	taskDao dao.Task,
	taskDaoV2 daov2.Task,
	logger telemetry.Logger,
	message entity.Message,
) *CreateMessageMutation {
	return &CreateMessageMutation{
		logger:           logger,
		stateSyncer:      stateSyncer,
		messageDao:       messageDao,
		messageDaoV2:     messageDaoV2,
		taskDao:          taskDao,
		taskDaoV2:        taskDaoV2,
		id:               stateSyncer.NextMutationID(),
		message:          message,
		notifierPrepared: false,
	}
}
