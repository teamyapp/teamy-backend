package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTaskMutation struct {
	dataCollector telemetry.DataCollector
	stateSyncer   *realtime.StateSyncer
	taskDao       dao.Task
	id            uint64
	task          entity.Task
}

func (c *CreateTaskMutation) GetID() uint64 {
	return c.id
}

func (c *CreateTaskMutation) Execute(ct context.Context) error {
	err := c.taskDao.CreateTask(ct, c.task)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return err
	}

	return nil
}

func (c *CreateTaskMutation) Undo() error {
	return nil
}

func (c *CreateTaskMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return c.stateSyncer.GetClientNotifiersByTeamID(ct, c.task.OwningTeamID)
}

func (c *CreateTaskMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TaskCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.task,
	}
}

func (c *CreateTaskMutation) CleanUp(ct context.Context) error {
	return nil
}

func NewCreateTaskMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	taskDao dao.Task,
	task entity.Task,
) *CreateTaskMutation {
	return &CreateTaskMutation{
		dataCollector: dataCollector,
		stateSyncer:   stateSyncer,
		taskDao:       taskDao,
		id:            stateSyncer.NextMutationID(),
		task:          task,
	}
}
