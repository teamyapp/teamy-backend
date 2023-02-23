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

type DeleteTaskMutation struct {
	dataCollector telemetry.DataCollector
	stateSyncer   *realtime.StateSyncer
	taskDao       dao.Task
	id            uint64
	task          entity.Task
}

var _ realtime.Mutation = (*DeleteTaskMutation)(nil)

func (d *DeleteTaskMutation) GetID() uint64 {
	return d.id
}

func (d *DeleteTaskMutation) ExecuteV2(ct context.Context, tx *sql.Tx) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (d *DeleteTaskMutation) PrepareClientNotifiers(ct context.Context, tx *sql.Tx) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (d *DeleteTaskMutation) Execute(ct context.Context) *errs.Error {
	err := d.taskDao.DeleteTask(ct, d.task.ID)
	if err != nil {
		d.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (d *DeleteTaskMutation) Undo() *errs.Error {
	return nil
}

func (d *DeleteTaskMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
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

func (d *DeleteTaskMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewDeleteTaskMutation(
	dataCollector telemetry.DataCollector,
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
