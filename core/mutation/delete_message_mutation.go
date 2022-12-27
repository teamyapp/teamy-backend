package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteMessageMutation struct {
	stateSyncer   *realtime.StateSyncer
	message       entity.Message
	messageDao    dao.Message
	dataCollector obs.DataCollector
	id            uint64
	taskDao       dao.Task
}

func (c *DeleteMessageMutation) GetID() uint64 {
	return c.id
}

func (d *DeleteMessageMutation) Execute(ct context.Context) error {
	err := d.messageDao.DeleteMessage(ct, d.message.ID)
	if err != nil {
		d.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (d *DeleteMessageMutation) Undo() error {
	return nil
}

func (d *DeleteMessageMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	task, err := d.taskDao.FindTaskByCommentsThreadID(ct, d.message.ThreadID)
	if err != nil {
		d.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return []*realtime.ClientNotifier{}, err
	}

	return d.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
}

func (d *DeleteMessageMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.MessageCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        d.message.ID,
	}
}

func NewDeleteMessageMutation(
	stateSyncer *realtime.StateSyncer,
	messageDao dao.Message,
	taskDao dao.Task,
	dataCollector obs.DataCollector,
	message entity.Message) *DeleteMessageMutation {
	return &DeleteMessageMutation{
		stateSyncer:   stateSyncer,
		messageDao:    messageDao,
		taskDao:       taskDao,
		dataCollector: dataCollector,
		id:            stateSyncer.NextMutationID(),
		message:       message,
	}
}
