package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteTaskAwaitForRelationMutation struct {
	stateSyncer             *realtime.StateSyncer
	taskAwaitForRelationDao dao.TaskAwaitForRelation
	dataCollector           obs.DataCollector
	id                      uint64
	awaitingTask            entity.Task
	awaitForTaskID          uint64
}

func (c *DeleteTaskAwaitForRelationMutation) GetID() uint64 {
	return c.id
}

func (d *DeleteTaskAwaitForRelationMutation) Execute(ct context.Context) error {
	err := d.taskAwaitForRelationDao.DeleteRelation(ct, d.awaitingTask.ID, d.awaitForTaskID)
	if err != nil {
		d.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (d *DeleteTaskAwaitForRelationMutation) Undo() error {
	return nil
}

func (d *DeleteTaskAwaitForRelationMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.awaitingTask.OwningTeamID)
}

func (d *DeleteTaskAwaitForRelationMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.TaskAwaitForRelationCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload: struct {
			AwaitingTaskID uint64
			AwaitForTaskID uint64
		}{
			AwaitingTaskID: d.awaitingTask.ID,
			AwaitForTaskID: d.awaitForTaskID,
		},
	}
}

func NewDeleteTaskAwaitForRelationMutation(
	stateSyncer *realtime.StateSyncer,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	dataCollector obs.DataCollector,
	awaitingTask entity.Task,
	awaitForTaskID uint64) *DeleteTaskAwaitForRelationMutation {
	return &DeleteTaskAwaitForRelationMutation{
		stateSyncer:             stateSyncer,
		taskAwaitForRelationDao: taskAwaitForRelationDao,
		dataCollector:           dataCollector,
		id:                      stateSyncer.NextMutationID(),
		awaitingTask:            awaitingTask,
		awaitForTaskID:          awaitForTaskID,
	}
}
