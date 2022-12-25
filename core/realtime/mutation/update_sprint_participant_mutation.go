package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateSprintParticipantMutation struct {
	id                   uint64
	teamID               uint64
	stateSyncer          *realtime.StateSyncer
	sprintParticipant    entity.SprintParticipant
	sprintParticipantDao dao.SprintParticipant
	dataCollector        obs.DataCollector
}

func (c *UpdateSprintParticipantMutation) GetID() uint64 {
	return c.id
}

func (u *UpdateSprintParticipantMutation) Execute(ct context.Context) error {
	err := u.sprintParticipantDao.UpdateSprintParticipant(ct, u.sprintParticipant)
	if err != nil {
		u.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	return nil
}

func (u *UpdateSprintParticipantMutation) Undo() error {
	return nil
}

func (u *UpdateSprintParticipantMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	return u.stateSyncer.GetClientNotifiersByTeamID(ct, u.teamID)
}

func (u *UpdateSprintParticipantMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.SprintParticipantCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.sprintParticipant,
	}
}

func NewUpdateSprintParticipantMutation(
	teamID uint64,
	stateSyncer *realtime.StateSyncer,
	sprintParticipant entity.SprintParticipant,
	sprintParticipantDao dao.SprintParticipant,
	dataCollector obs.DataCollector) *UpdateSprintParticipantMutation {
	return &UpdateSprintParticipantMutation{
		id:                   stateSyncer.NextMutationID(),
		teamID:               teamID,
		stateSyncer:          stateSyncer,
		sprintParticipant:    sprintParticipant,
		sprintParticipantDao: sprintParticipantDao,
		dataCollector:        dataCollector,
	}
}
