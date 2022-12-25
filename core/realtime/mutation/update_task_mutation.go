package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateTaskMutation struct {
	id            uint64
	teamID        uint64
	stateSyncer   *realtime.StateSyncer
	task          entity.Task
	taskDao       dao.Task
	dataCollector obs.DataCollector
}

func (c *UpdateTaskMutation) GetID() uint64 {
	return c.id
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
	return u.stateSyncer.GetClientNotifiersByTeamID(ct, u.teamID)
}

func (u *UpdateTaskMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.TaskCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.task,
	}
}

func NewUpdateTaskMutation(
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	task entity.Task,
	taskDao dao.Task,
	dataCollector obs.DataCollector) *UpdateTaskMutation {
	return &UpdateTaskMutation{
		id:            stateSyncer.NextMutationID(),
		teamID:        teamID,
		stateSyncer:   stateSyncer,
		task:          task,
		taskDao:       taskDao,
		dataCollector: dataCollector,
	}
}
