package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteSprintParticipantMutation struct {
	id                   uint64
	teamID               uint64
	stateSyncer          *realtime.StateSyncer
	userID               uint64
	sprintID             uint64
	sprintParticipantDao dao.SprintParticipant
	dataCollector        obs.DataCollector
}

func (c *DeleteSprintParticipantMutation) GetID() uint64 {
	return c.id
}

func (d *DeleteSprintParticipantMutation) Execute(ct context.Context) error {
	err := d.sprintParticipantDao.DeleteSprintParticipant(ct, d.sprintID, d.userID)
	if err != nil {
		d.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (d *DeleteSprintParticipantMutation) Undo() error {
	return nil
}

func (d *DeleteSprintParticipantMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return d.stateSyncer.GetClientNotifiersByTeamID(ct, d.teamID)
}

func (d *DeleteSprintParticipantMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             d.id,
		CollectionType: realtime.SprintParticipantCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload: struct {
			SprintID uint64
			UserID   uint64
		}{
			SprintID: d.sprintID,
			UserID:   d.userID,
		},
	}
}

func NewDeleteSprintParticipantMutation(
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	userID uint64,
	sprintID uint64,
	sprintParticipantDao dao.SprintParticipant,
	dataCollector obs.DataCollector) *DeleteSprintParticipantMutation {
	return &DeleteSprintParticipantMutation{
		id:                   stateSyncer.NextMutationID(),
		teamID:               teamID,
		stateSyncer:          stateSyncer,
		userID:               userID,
		sprintID:             sprintID,
		sprintParticipantDao: sprintParticipantDao,
		dataCollector:        dataCollector,
	}
}
