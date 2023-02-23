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

type UpdateTaskMutation struct {
	dataCollector telemetry.DataCollector
	stateSyncer   *realtime.StateSyncer
	taskDao       dao.Task
	id            uint64
	task          entity.Task
}

var _ realtime.Mutation = (*UpdateTaskMutation)(nil)

func (u *UpdateTaskMutation) GetID() uint64 {
	return u.id
}

func (u *UpdateTaskMutation) ExecuteV2(ct context.Context, tx *sql.Tx) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (u *UpdateTaskMutation) PrepareClientNotifiers(ct context.Context, tx *sql.Tx) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (u *UpdateTaskMutation) Execute(ct context.Context) *errs.Error {
	err := u.taskDao.UpdateTask(ct, u.task)
	if err != nil {
		u.dataCollector.Logger.ErrorWithContext(ct, err)
		return err
	}

	return nil
}

func (u *UpdateTaskMutation) Undo() *errs.Error {
	return nil
}

func (u *UpdateTaskMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	return u.stateSyncer.GetClientNotifiersByTeamID(ct, u.task.OwningTeamID)
}

func (u *UpdateTaskMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.TaskCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.task,
	}
}

func (u *UpdateTaskMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewUpdateTaskMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	taskDao dao.Task,
	task entity.Task,
) *UpdateTaskMutation {
	return &UpdateTaskMutation{
		dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		taskDao:       taskDao,
		id:            stateSyncer.NextMutationID(),
		task:          task,
	}
}
