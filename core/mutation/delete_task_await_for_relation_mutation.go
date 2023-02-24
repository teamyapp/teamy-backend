package mutation

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteTaskAwaitForRelationMutation struct {
	dataCollector           telemetry.DataCollector
	stateSyncer             *realtime.StateSyncer
	taskAwaitForRelationDao dao.TaskAwaitForRelation
	id                      uint64
	awaitingTask            entity.Task
	awaitForTaskID          uint64
}

var _ realtime.Mutation = (*DeleteTaskAwaitForRelationMutation)(nil)

func (d *DeleteTaskAwaitForRelationMutation) GetID() uint64 {
	return d.id
}

func (d *DeleteTaskAwaitForRelationMutation) ExecuteV2(ct context.Context, tx *sql.Tx) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (d *DeleteTaskAwaitForRelationMutation) PrepareClientNotifiers(ct context.Context, tx *sql.Tx) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (d *DeleteTaskAwaitForRelationMutation) Execute(ct context.Context) *errs.Error {
	err := d.taskAwaitForRelationDao.DeleteRelation(ct, d.awaitingTask.ID, d.awaitForTaskID)
	if err != nil {
		d.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (d *DeleteTaskAwaitForRelationMutation) Undo() *errs.Error {
	return nil
}

func (d *DeleteTaskAwaitForRelationMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.awaitingTask.OwningTeamID)
}

func (d *DeleteTaskAwaitForRelationMutation) GetClientNotifiersV2() []*realtime.ClientNotifier {
	//TODO implement me
	panic("implement me")
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

func (d *DeleteTaskAwaitForRelationMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewDeleteTaskAwaitForRelationMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	awaitingTask entity.Task,
	awaitForTaskID uint64,
) *DeleteTaskAwaitForRelationMutation {
	return &DeleteTaskAwaitForRelationMutation{
		dataCollector:           dataCollector,
		stateSyncer:             stateSyncer,
		taskAwaitForRelationDao: taskAwaitForRelationDao,
		id:                      stateSyncer.NextMutationID(),
		awaitingTask:            awaitingTask,
		awaitForTaskID:          awaitForTaskID,
	}
}
