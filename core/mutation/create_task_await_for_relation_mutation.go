package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTaskAwaitForRelationMutation struct {
	stateSyncer             *realtime.StateSyncer
	taskAwaitForRelationDao dao.TaskAwaitForRelation
	taskDao                 dao.Task
	dataCollector           obs.DataCollector
	id                      uint64
	taskAwaitForRelation    entity.TaskAwaitForRelation
}

func (c *CreateTaskAwaitForRelationMutation) GetID() uint64 {
	return c.id
}

func (c *CreateTaskAwaitForRelationMutation) Execute(ct context.Context) error {
	err := c.taskAwaitForRelationDao.CreateRelation(ct, c.taskAwaitForRelation)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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

func NewCreateTaskAwaitForRelationMutation(
	stateSyncer *realtime.StateSyncer,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	taskDao dao.Task,
	dataCollector obs.DataCollector,
	taskAwaitForRelation entity.TaskAwaitForRelation) *CreateTaskAwaitForRelationMutation {
	return &CreateTaskAwaitForRelationMutation{
		stateSyncer:             stateSyncer,
		taskAwaitForRelationDao: taskAwaitForRelationDao,
		taskDao:                 taskDao,
		dataCollector:           dataCollector,
		id:                      stateSyncer.NextMutationID(),
		taskAwaitForRelation:    taskAwaitForRelation,
	}
}
