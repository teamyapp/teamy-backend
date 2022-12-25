package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type CreateSprintParticipantMutation struct {
	id                   uint64
	teamID               uint64
	stateSyncer          *realtime.StateSyncer
	sprintParticipant    entity.SprintParticipant
	sprintParticipantDao dao.SprintParticipant
	dataCollector        obs.DataCollector
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
	return c.stateSyncer.GetClientNotifiersByTeamID(ct, c.teamID)
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
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	sprintParticipant entity.SprintParticipant,
	sprintParticipantDao dao.SprintParticipant,
	dataCollector obs.DataCollector) *CreateSprintParticipantMutation {
	return &CreateSprintParticipantMutation{
		id:                   stateSyncer.NextMutationID(),
		teamID:               teamID,
		stateSyncer:          stateSyncer,
		sprintParticipant:    sprintParticipant,
		sprintParticipantDao: sprintParticipantDao,
		dataCollector:        dataCollector,
	}
}
