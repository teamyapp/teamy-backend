package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateTaskMutation struct {
	dataCollector obs.DataCollector
	stateSyncer   *realtime.StateSyncer
	taskDao       dao.Task
	id            uint64
	task          entity.Task
}

func (u *UpdateTaskMutation) GetID() uint64 {
	return u.id
}

func (u *UpdateTaskMutation) Execute(ct context.Context) error {
	err := u.taskDao.UpdateTask(ct, u.task)
	if err != nil {
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (u *UpdateTaskMutation) Undo() error {
	return nil
}

func (u *UpdateTaskMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
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

func (u *UpdateTaskMutation) CleanUp(ct context.Context) error {
	return nil
}

func NewUpdateTaskMutation(
	dataCollector obs.DataCollector,
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
