package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateSprintTaskRelationMutation struct {
	id                    uint64
	teamID                uint64
	stateSyncer           *realtime.StateSyncer
	sprintTaskRelation    entity.SprintTaskRelation
	sprintTaskRelationDao dao.SprintTaskRelation
	dataCollector         obs.DataCollector
}

func (c *CreateSprintTaskRelationMutation) GetID() uint64 {
	return c.id
}

func (c *CreateSprintTaskRelationMutation) Execute(ct context.Context) error {
	err := c.sprintTaskRelationDao.CreateSprintTaskRelation(ct, c.sprintTaskRelation)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (c *CreateSprintTaskRelationMutation) Undo() error {
	return nil
}

func (c *CreateSprintTaskRelationMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return c.stateSyncer.GetClientNotifiersByTeamID(ct, c.teamID)
}

func (c *CreateSprintTaskRelationMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.SprintTaskRelationCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.sprintTaskRelation,
	}
}

func NewCreateSprintTaskRelationMutation(
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	sprintTaskRelation entity.SprintTaskRelation,
	sprintTaskRelationDao dao.SprintTaskRelation,
	dataCollector obs.DataCollector) *CreateSprintTaskRelationMutation {
	return &CreateSprintTaskRelationMutation{
		id:                    stateSyncer.NextMutationID(),
		teamID:                teamID,
		stateSyncer:           stateSyncer,
		sprintTaskRelation:    sprintTaskRelation,
		sprintTaskRelationDao: sprintTaskRelationDao,
		dataCollector:         dataCollector,
	}
}
