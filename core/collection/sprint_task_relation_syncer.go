package collection

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type SprintTaskRelationSyncer struct {
	dataCollector         obs.DataCollector
	realTimeStateSyncer   *realtime.StateSyncer
	sprintTaskRelationDao dao.SprintTaskRelation
	sprintDao             dao.Sprint
}

func (s SprintTaskRelationSyncer) CreateAndSyncSprintTaskRelation(
	ct context.Context,
	tx realtime.Transaction,
	sprintTaskRelation entity.SprintTaskRelation,
) error {
	err := s.sprintTaskRelationDao.CreateSprintTaskRelation(ct, sprintTaskRelation)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.SprintTaskRelationCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        sprintTaskRelation,
	})
	return nil
}

func (s SprintTaskRelationSyncer) DeleteAndSyncSprintTaskRelation(
	ct context.Context,
	tx realtime.Transaction,
	sprintID uint64,
	taskID uint64,
) error {
	err := s.sprintTaskRelationDao.DeleteSprintTaskRelation(ct, sprintID, taskID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.SprintTaskRelationCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload: struct {
			SprintID uint64
			TaskID   uint64
		}{
			SprintID: sprintID,
			TaskID:   taskID,
		},
	})
	return nil
}

func NewSprintTaskRelationSyncer(
	dataCollector obs.DataCollector,
	realTimeStateSyncer *realtime.StateSyncer,
	sprintTaskRelationDao dao.SprintTaskRelation,
	sprintDao dao.Sprint,
) SprintTaskRelationSyncer {
	return SprintTaskRelationSyncer{
		dataCollector:         dataCollector,
		realTimeStateSyncer:   realTimeStateSyncer,
		sprintTaskRelationDao: sprintTaskRelationDao,
		sprintDao:             sprintDao,
	}
}
