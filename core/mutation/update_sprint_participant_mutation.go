package mutation

import (
	"context"

	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type UpdateSprintParticipantMutation struct {
	dataCollector        telemetry.DataCollector
	stateSyncer          *realtime.StateSyncer
	sprintParticipantDao dao.SprintParticipant
	sprintDao            dao.Sprint
	id                   uint64
	sprintParticipant    entity.SprintParticipant
}

func (u *UpdateSprintParticipantMutation) GetID() uint64 {
	return u.id
}

func (u *UpdateSprintParticipantMutation) Execute(ct context.Context) error {
	err := u.sprintParticipantDao.UpdateSprintParticipant(ct, u.sprintParticipant)
	if err != nil {
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return err
	}

	return nil
}

func (u *UpdateSprintParticipantMutation) Undo() error {
	return nil
}

func (u *UpdateSprintParticipantMutation) GetClientNotifiers(ct context.Context) ([]*realtime.ClientNotifier, error) {
	sprint, err := u.sprintDao.FindSprintByID(ct, u.sprintParticipant.SprintID)
	if err != nil {
		u.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return []*realtime.ClientNotifier{}, err
	}

	return u.stateSyncer.GetClientNotifiersByTeamID(ct, sprint.OwningTeamID)
}

func (u *UpdateSprintParticipantMutation) ToMessage() realtime.MutationMessage {
	return realtime.MutationMessage{
		ID:             u.id,
		CollectionType: realtime.SprintParticipantCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        u.sprintParticipant,
	}
}

func (u *UpdateSprintParticipantMutation) CleanUp(ct context.Context) error {
	return nil
}

func NewUpdateSprintParticipantMutation(
	dataCollector telemetry.DataCollector,
	stateSyncer *realtime.StateSyncer,
	sprintParticipantDao dao.SprintParticipant,
	sprintDao dao.Sprint,
	sprintParticipant entity.SprintParticipant,
) *UpdateSprintParticipantMutation {
	return &UpdateSprintParticipantMutation{
		dataCollector:        dataCollector,
		stateSyncer:          stateSyncer,
		sprintParticipantDao: sprintParticipantDao,
		sprintDao:            sprintDao,
		id:                   stateSyncer.NextMutationID(),
		sprintParticipant:    sprintParticipant,
	}
}
