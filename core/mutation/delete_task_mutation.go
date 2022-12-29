package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteTaskMutation struct {
        dataCollector obs.DataCollector
	stateSyncer   *realtime.StateSyncer
	taskDao       dao.Task
	id            uint64
	task          entity.Task
}

func (c *DeleteTaskMutation) GetID() uint64 {
	return c.id
}

func (d *DeleteTaskMutation) Execute(ct context.Context) error {
	err := d.taskDao.DeleteTask(ct, d.task.ID)
	if err != nil {
		d.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (d *DeleteTaskMutation) Undo() error {
	return nil
}

func (d *DeleteTaskMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.task.OwningTeamID)
}

func (d *DeleteTaskMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.TaskCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        d.task.ID,
	}
}

func NewDeleteTaskMutation(
        dataCollector obs.DataCollector,
	stateSyncer *realtime.StateSyncer,
	taskDao dao.Task,
	task entity.Task,
) *DeleteTaskMutation {
	return &DeleteTaskMutation{
	        dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		taskDao:       taskDao,	
		id:            stateSyncer.NextMutationID(),
		task:          task,
	}
}
