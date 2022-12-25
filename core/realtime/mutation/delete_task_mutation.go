package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteTaskMutation struct {
	id            uint64
	teamID        uint64
	stateSyncer   *realtime.StateSyncer
	taskID        uint64
	taskDao       dao.Task
	dataCollector obs.DataCollector
}

func (c *DeleteTaskMutation) GetID() uint64 {
	return c.id
}

func (d *DeleteTaskMutation) Execute(ct context.Context) error {
	err := d.taskDao.DeleteTask(ct, d.taskID)
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
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.teamID)
}

func (d *DeleteTaskMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.TaskCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        d.taskID,
	}
}

func NewDeleteTaskMutation(
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	taskID uint64,
	taskDao dao.Task,
	dataCollector obs.DataCollector) *DeleteTaskMutation {
	return &DeleteTaskMutation{
		id:            stateSyncer.NextMutationID(),
		teamID:        teamID,
		stateSyncer:   stateSyncer,
		taskID:        taskID,
		taskDao:       taskDao,
		dataCollector: dataCollector,
	}
}
