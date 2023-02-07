package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateSprintTaskRelationMutation struct {
	dataCollector         telemetry.DataCollector
	stateSyncer           *realtime.StateSyncer
	sprintTaskRelationDao dao.SprintTaskRelation
	sprintDao             dao.Sprint
	id                    uint64
	sprintTaskRelation    entity.SprintTaskRelation
}

func (c *CreateSprintTaskRelationMutation) GetID() uint64 {
	return c.id
}

func (c *CreateSprintTaskRelationMutation) Execute(ct context.Context) *errs.Error {
	err := c.sprintTaskRelationDao.CreateSprintTaskRelation(ct, c.sprintTaskRelation)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return err
	}

	return nil
}

func (c *CreateSprintTaskRelationMutation) Undo() *errs.Error {
	return nil
}

func (c *CreateSprintTaskRelationMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, *errs.Error) {
	sprint, err := c.sprintDao.FindSprintByID(ct, c.sprintTaskRelation.SprintID)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return []*realtime.ClientNotifier{}, err
	}

	return c.stateSyncer.GetClientNotifiersByTeamID(ct, sprint.OwningTeamID)
}

func (c *CreateSprintTaskRelationMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.SprintTaskRelationCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.sprintTaskRelation,
	}
}

func (c *CreateSprintTaskRelationMutation) CleanUp(ct context.Context) *errs.Error {
	return nil
}

func NewCreateSprintTaskRelationMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	sprintTaskRelationDao dao.SprintTaskRelation,
	sprintDao dao.Sprint,
	sprintTaskRelation entity.SprintTaskRelation,
) *CreateSprintTaskRelationMutation {
	return &CreateSprintTaskRelationMutation{
		stateSyncer:           stateSyncer,
		sprintTaskRelationDao: sprintTaskRelationDao,
		sprintDao:             sprintDao,
		dataCollector:         dataCollector,
		id:                    stateSyncer.NextMutationID(),
		sprintTaskRelation:    sprintTaskRelation,
	}
}
