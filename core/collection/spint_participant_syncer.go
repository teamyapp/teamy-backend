package collection

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type SprintParticipantSyncer struct {
	dataCollector        obs.DataCollector
	realTimeStateSyncer  *realtime.StateSyncer
	sprintParticipantDao dao.SprintParticipant
	sprintDao            dao.Sprint
}

func (s SprintParticipantSyncer) CreateAndSyncSprintParticipant(
	ct context.Context,
	tx realtime.Transaction,
	sprintParticipant entity.SprintParticipant,
) error {
	err := s.sprintParticipantDao.CreateSprintParticipant(ct, sprintParticipant)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.SprintParticipantCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        sprintParticipant,
	})
	return nil
}

func (s SprintParticipantSyncer) UpdateAndSyncSprintParticipant(
	ct context.Context,
	tx realtime.Transaction,
	sprintParticipant entity.SprintParticipant,
) error {
	err := s.sprintParticipantDao.UpdateSprintParticipant(ct, sprintParticipant)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.SprintParticipantCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        sprintParticipant,
	})
	return nil
}

func (s SprintParticipantSyncer) DeleteAndSyncSprintParticipant(ct context.Context, tx realtime.Transaction, sprintID uint64, userID uint64) error {
	err := s.sprintParticipantDao.DeleteSprintParticipant(ct, sprintID, userID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.SprintParticipantCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload: struct {
			SprintID uint64
			UserID   uint64
		}{
			SprintID: sprintID,
			UserID:   userID,
		},
	})
	return nil
}

func NewSprintParticipantSyncer(
	dataCollector obs.DataCollector,
	realTimeStateSyncer *realtime.StateSyncer,
	sprintParticipantDao dao.SprintParticipant,
	sprintDao dao.Sprint,
) SprintParticipantSyncer {
	return SprintParticipantSyncer{
		dataCollector:        dataCollector,
		realTimeStateSyncer:  realTimeStateSyncer,
		sprintParticipantDao: sprintParticipantDao,
		sprintDao:            sprintDao,
	}
}
