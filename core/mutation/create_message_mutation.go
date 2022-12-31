package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateMessageMutation struct {
	dataCollector obs.DataCollector
	stateSyncer   *realtime.StateSyncer
	messageDao    dao.Message
	taskDao       dao.Task
	id            uint64
	message       entity.Message
}

func (c *CreateMessageMutation) GetID() uint64 {
	return c.id
}

func (c *CreateMessageMutation) Execute(ct context.Context) error {
	err := c.messageDao.CreateMessage(ct, c.message)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (c *CreateMessageMutation) Undo() error {
	return nil
}

func (c *CreateMessageMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	task, err := c.taskDao.FindTaskByCommentsThreadID(ct, c.message.ThreadID)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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

func NewCreateMessageMutation(
	stateSyncer *realtime.StateSyncer,
	messageDao dao.Message,
	taskDao dao.Task,
	dataCollector obs.DataCollector,
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
