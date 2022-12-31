package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type DeleteSprintParticipantMutation struct {
	dataCollector        obs.DataCollector
	stateSyncer          *realtime.StateSyncer
	sprintParticipantDao dao.SprintParticipant
	sprintDao            dao.Sprint
	id                   uint64
	userID               uint64
	sprintID             uint64
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
	sprint, err := d.sprintDao.FindSprintByID(ct, d.sprintID)
	if err != nil {
		d.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return []*realtime.ClientNotifier{}, err
	}

	return d.stateSyncer.GetClientNotifiersByTeamID(ct, sprint.OwningTeamID)
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

func (d *DeleteSprintParticipantMutation) CleanUp(ct context.Context) error {
	return nil
}

func NewDeleteSprintParticipantMutation(
	dataCollector obs.DataCollector,
	stateSyncer *realtime.StateSyncer,
	sprintParticipantDao dao.SprintParticipant,
	sprintDao dao.Sprint,
	userID uint64,
	sprintID uint64,
) *DeleteSprintParticipantMutation {
	return &DeleteSprintParticipantMutation{
		dataCollector:        dataCollector,
		stateSyncer:          stateSyncer,
		sprintParticipantDao: sprintParticipantDao,
		sprintDao:            sprintDao,
		id:                   stateSyncer.NextMutationID(),
		userID:               userID,
		sprintID:             sprintID,
	}
}
