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
}

func (s SprintTaskRelationSyncer) CreateAndSyncSprintTaskRelation(
	ct context.Context,
	sprintTaskRelation entity.SprintTaskRelation,
	OwningTeamID uint64,
) error {
	err := s.sprintTaskRelationDao.CreateSprintTaskRelation(ct, sprintTaskRelation)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	s.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.SprintTaskRelationCollectionType,
		MutationType:   realtime.CreateMutationType,
		TeamIDs: []uint64{
			OwningTeamID,
		},
		Payload: sprintTaskRelation,
	})
	return nil
}

func (s SprintTaskRelationSyncer) DeleteAndSyncSprintTaskRelation(
	ct context.Context,
	sprintID uint64,
	taskID uint64,
	OwningTeamID uint64,
) error {
	err := s.sprintTaskRelationDao.DeleteSprintTaskRelation(ct, sprintID, taskID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	s.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.SprintTaskRelationCollectionType,
		MutationType:   realtime.DeleteMutationType,
		TeamIDs: []uint64{
			OwningTeamID,
		},
		Payload: realtime.DeleteSprintTaskRelationPayload{SprintID: sprintID, TaskID: taskID},
	})
	return nil

}

func NewSprintTaskRelationSyncer(
	dataCollector obs.DataCollector,
	realTimeStateSyncer *realtime.StateSyncer,
	sprintTaskRelationDao dao.SprintTaskRelation,
) SprintTaskRelationSyncer {
	return SprintTaskRelationSyncer{
		dataCollector:         dataCollector,
		realTimeStateSyncer:   realTimeStateSyncer,
		sprintTaskRelationDao: sprintTaskRelationDao,
	}
}
