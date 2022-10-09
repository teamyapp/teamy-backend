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
	sprintTaskRelation entity.SprintTaskRelation,
) error {
	err := s.sprintTaskRelationDao.CreateSprintTaskRelation(ct, sprintTaskRelation)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	sprint, err := s.sprintDao.FindSprintByID(ct, sprintTaskRelation.SprintID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	s.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.SprintTaskRelationCollectionType,
		MutationType:   realtime.CreateMutationType,
		TeamIDs: []uint64{
			sprint.OwningTeamID,
		},
		Payload: sprintTaskRelation,
	})
	return nil
}

func (s SprintTaskRelationSyncer) DeleteAndSyncSprintTaskRelation(
	ct context.Context,
	sprintID uint64,
	taskID uint64,
) error {
	err := s.sprintTaskRelationDao.DeleteSprintTaskRelation(ct, sprintID, taskID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	sprint, err := s.sprintDao.FindSprintByID(ct, sprintID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	s.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.SprintTaskRelationCollectionType,
		MutationType:   realtime.DeleteMutationType,
		TeamIDs: []uint64{
			sprint.OwningTeamID,
		},
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
