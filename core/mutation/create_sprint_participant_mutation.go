package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateSprintParticipantMutation struct {
	dataCollector        obs.DataCollector
	stateSyncer          *realtime.StateSyncer
	sprintParticipantDao dao.SprintParticipant
	sprintDao            dao.Sprint
	id                   uint64
	sprintParticipant    entity.SprintParticipant
}

func (c *CreateSprintParticipantMutation) GetID() uint64 {
	return c.id
}

func (c *CreateSprintParticipantMutation) Execute(ct context.Context) error {
	err := c.sprintParticipantDao.CreateSprintParticipant(ct, c.sprintParticipant)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (c *CreateSprintParticipantMutation) Undo() error {
	return nil
}

func (c *CreateSprintParticipantMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	sprint, err := c.sprintDao.FindSprintByID(ct, c.sprintParticipant.SprintID)
	if err != nil {
		c.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return []*realtime.ClientNotifier{}, err
	}

	return c.stateSyncer.GetClientNotifiersByTeamID(ct, sprint.OwningTeamID)
}

func (c *CreateSprintParticipantMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             c.id,
		CollectionType: realtime.SprintParticipantCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        c.sprintParticipant,
	}
}

func NewCreateSprintParticipantMutation(
        dataCollector obs.DataCollector,
	stateSyncer *realtime.StateSyncer,
	sprintParticipantDao dao.SprintParticipant,
	sprintDao dao.Sprint,
	sprintParticipant entity.SprintParticipant,
) *CreateSprintParticipantMutation {
	return &CreateSprintParticipantMutation{
		dataCollector:        dataCollector,
		stateSyncer:          stateSyncer,
		sprintParticipantDao: sprintParticipantDao,
		sprintDao:            sprintDao,
		id:                   stateSyncer.NextMutationID(),
		sprintParticipant:    sprintParticipant,
	}
}
