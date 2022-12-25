package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTaskMutation struct {
	id            uint64
	teamID        uint64
	stateSyncer   *realtime.StateSyncer
	task          entity.Task
	taskDao       dao.Task
	dataCollector obs.DataCollector
}

func (c *CreateTaskMutation) GetID() uint64 {
	return c.id
}

func (c *CreateTaskMutation) Execute(ct context.Context) error {
	err := c.taskDao.CreateTask(ct, c.task)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (c *CreateTaskMutation) Undo() error {
	return nil
}

func (c *CreateTaskMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return c.stateSyncer.GetClientNotifiersByTeamID(ct, c.teamID)
}

func (c *CreateTaskMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TaskCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.task,
	}
}

func NewCreateTaskMutation(
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	task entity.Task,
	taskDao dao.Task,
	dataCollector obs.DataCollector) *CreateTaskMutation {
	return &CreateTaskMutation{
		id:            stateSyncer.NextMutationID(),
		teamID:        teamID,
		stateSyncer:   stateSyncer,
		task:          task,
		taskDao:       taskDao,
		dataCollector: dataCollector,
	}
}
