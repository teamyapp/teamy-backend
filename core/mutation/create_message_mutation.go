package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateMessageMutation struct {
	dataCollector telemetry.DataCollector
	stateSyncer   *realtime.StateSyncer
	messageDao    dao.Message
	taskDao       dao.Task
	id            uint64
	message       entity.Message
}

func (c *CreateMessageMutation) GetID() uint64 {
	return c.id
}

func (c *CreateMessageMutation) Execute(ct context.Context) *errs.Error {
	err := c.messageDao.CreateMessage(ct, c.message)
	if err != nil {
		c.dataCollector.Logger.ErrorWithContext(ct, err)
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
		c.dataCollector.Logger.ErrorWithContext(ct, err)
		return []*realtime.ClientNotifier{}, err
	}

	return c.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
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
	taskDao dao.Task,
	dataCollector telemetry.DataCollector,
	message entity.Message,
) *CreateMessageMutation {
	return &CreateMessageMutation{
		dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		messageDao:    messageDao,
		taskDao:       taskDao,
		id:            stateSyncer.NextMutationID(),
		message:       message,
	}
}
