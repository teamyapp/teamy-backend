package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteTaskAwaitForRelationMutation struct {
	id                      uint64
	teamID                  uint64
	stateSyncer             *realtime.StateSyncer
	awaitingTaskID          uint64
	awaitForTaskID          uint64
	taskAwaitForRelationDao dao.TaskAwaitForRelation
	dataCollector           obs.DataCollector
}

func (c *DeleteTaskAwaitForRelationMutation) GetID() uint64 {
	return c.id
}

func (d *DeleteTaskAwaitForRelationMutation) Execute(ct context.Context) error {
	err := d.taskAwaitForRelationDao.DeleteRelation(ct, d.awaitingTaskID, d.awaitForTaskID)
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
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.teamID)
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
			AwaitingTaskID: d.awaitingTaskID,
			AwaitForTaskID: d.awaitForTaskID,
		},
	}
}

func NewDeleteTaskAwaitForRelationMutation(
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	awaitingTaskID uint64,
	awaitForTaskID uint64,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	dataCollector obs.DataCollector) *DeleteTaskAwaitForRelationMutation {
	return &DeleteTaskAwaitForRelationMutation{
		id:             stateSyncer.NextMutationID(),
		teamID:         teamID,
		stateSyncer:    stateSyncer,
		awaitingTaskID: awaitingTaskID,
		awaitForTaskID: awaitForTaskID,
		dataCollector:  dataCollector,
	}
}
