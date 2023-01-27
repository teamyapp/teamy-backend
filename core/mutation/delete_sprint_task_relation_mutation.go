package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteSprintTaskRelationMutation struct {
	dataCollector         obs.DataCollector
	stateSyncer           *realtime.StateSyncer
	sprintTaskRelationDao dao.SprintTaskRelation
	id                    uint64
	sprintID              uint64
	task                  entity.Task
}

func (d *DeleteSprintTaskRelationMutation) GetID() uint64 {
	return d.id
}

func (d *DeleteSprintTaskRelationMutation) Execute(ct context.Context) error {
	err := d.sprintTaskRelationDao.DeleteSprintTaskRelation(ct, d.sprintID, d.task.ID)
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
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.task.OwningTeamID)
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
			TaskID:   d.task.ID,
		},
	}
}

func (d *DeleteSprintTaskRelationMutation) CleanUp(ct context.Context) error {
	return nil
}

func NewDeleteSprintTaskRelationMutation(
	dataCollector obs.DataCollector,
	stateSyncer *realtime.StateSyncer,
	sprintTaskRelationDao dao.SprintTaskRelation,
	sprintID uint64,
	task entity.Task,
) *DeleteSprintTaskRelationMutation {
	return &DeleteSprintTaskRelationMutation{
		dataCollector:         dataCollector,
		stateSyncer:           stateSyncer,
		sprintTaskRelationDao: sprintTaskRelationDao,
		id:                    stateSyncer.NextMutationID(),
		sprintID:              sprintID,
		task:                  task,
	}
}
