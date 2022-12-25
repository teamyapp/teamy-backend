package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteSprintTaskRelationMutation struct {
	id                    uint64
	teamID                uint64
	stateSyncer           *realtime.StateSyncer
	sprintID              uint64
	taskID                uint64
	sprintTaskRelationDao dao.SprintTaskRelation
	dataCollector         obs.DataCollector
}

func (c *DeleteSprintTaskRelationMutation) GetID() uint64 {
	return c.id
}

func (d *DeleteSprintTaskRelationMutation) Execute(ct context.Context) error {
	err := d.sprintTaskRelationDao.DeleteSprintTaskRelation(ct, d.sprintID, d.taskID)
	if err != nil {
		d.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (d *DeleteSprintTaskRelationMutation) Undo() error {
	return nil
}

func (d *DeleteSprintTaskRelationMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.teamID)
}

func (d *DeleteSprintTaskRelationMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.SprintTaskRelationCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload: struct {
			SprintID uint64
			TaskID   uint64
		}{
			SprintID: d.sprintID,
			TaskID:   d.taskID,
		},
	}
}

func NewDeleteSprintTaskRelationMutation(
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	sprintID uint64,
	taskID uint64,
	dataCollector obs.DataCollector) *DeleteSprintTaskRelationMutation {
	return &DeleteSprintTaskRelationMutation{
		id:            stateSyncer.NextMutationID(),
		teamID:        teamID,
		stateSyncer:   stateSyncer,
		sprintID:      sprintID,
		taskID:        taskID,
		dataCollector: dataCollector,
	}
}
