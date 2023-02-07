package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteSprintTaskRelationMutation struct {
	dataCollector         telemetry.DataCollector
	stateSyncer           *realtime.StateSyncer
	sprintTaskRelationDao dao.SprintTaskRelation
	id                    uint64
	sprintID              uint64
	task                  entity.Task
}

func (d *DeleteSprintTaskRelationMutation) GetID() uint64 {
	return d.id
}

func (d *DeleteSprintTaskRelationMutation) Execute(ct context.Context) *errs.Error {
	err := d.sprintTaskRelationDao.DeleteSprintTaskRelation(ct, d.sprintID, d.task.ID)
	if err != nil {
		d.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return err
	}

	return nil
}

func (d *DeleteSprintTaskRelationMutation) Undo() *errs.Error {
	return nil
}

func (d *DeleteSprintTaskRelationMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
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

func (d *DeleteSprintTaskRelationMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewDeleteSprintTaskRelationMutation(
	dataCollector telemetry.DataCollector,
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
