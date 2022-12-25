package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateTaskAwaitForRelationMutation struct {
	id                      uint64
	teamID                  uint64
	stateSyncer             *realtime.StateSyncer
	taskAwaitForRelation    entity.TaskAwaitForRelation
	taskAwaitForRelationDao dao.TaskAwaitForRelation
	dataCollector           obs.DataCollector
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
	return c.stateSyncer.GetClientNotifiersByTeamID(ct, c.teamID)
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
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	taskAwaitForRelation entity.TaskAwaitForRelation,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	dataCollector obs.DataCollector) *CreateTaskAwaitForRelationMutation {
	return &CreateTaskAwaitForRelationMutation{
		id:                      stateSyncer.NextMutationID(),
		teamID:                  teamID,
		stateSyncer:             stateSyncer,
		taskAwaitForRelation:    taskAwaitForRelation,
		taskAwaitForRelationDao: taskAwaitForRelationDao,
		dataCollector:           dataCollector,
	}
}
