package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTaskAwaitForRelationMutation struct {
	dataCollector           telemetry.DataCollector
	stateSyncer             *realtime.StateSyncer
	taskAwaitForRelationDao dao.TaskAwaitForRelation
	taskDao                 dao.Task
	id                      uint64
	taskAwaitForRelation    entity.TaskAwaitForRelation
}

func (c *CreateTaskAwaitForRelationMutation) GetID() uint64 {
	return c.id
}

func (c *CreateTaskAwaitForRelationMutation) Execute(ct context.Context) error {
	err := c.taskAwaitForRelationDao.CreateRelation(ct, c.taskAwaitForRelation)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return err
	}

	return nil
}

func (c *CreateTaskAwaitForRelationMutation) Undo() error {
	return nil
}

func (c *CreateTaskAwaitForRelationMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	task, err := c.taskDao.FindTaskByID(ct, c.taskAwaitForRelation.AwaitForTaskID)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return []*realtime.ClientNotifier{}, err
	}

	return c.stateSyncer.GetClientNotifiersByTeamID(ct, task.OwningTeamID)
}

func (c *CreateTaskAwaitForRelationMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.TaskAwaitForRelationCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.taskAwaitForRelation,
	}
}

func (c *CreateTaskAwaitForRelationMutation) CleanUp(ct context.Context) error {
	return nil
}

func NewCreateTaskAwaitForRelationMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	taskDao dao.Task,
	taskAwaitForRelation entity.TaskAwaitForRelation,
) *CreateTaskAwaitForRelationMutation {
	return &CreateTaskAwaitForRelationMutation{
		dataCollector:           dataCollector,
		stateSyncer:             stateSyncer,
		taskAwaitForRelationDao: taskAwaitForRelationDao,
		taskDao:                 taskDao,
		id:                      stateSyncer.NextMutationID(),
		taskAwaitForRelation:    taskAwaitForRelation,
	}
}
