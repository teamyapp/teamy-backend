package collection

import (
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type SprintTaskRelationSyncer struct {
	realTimeStateSyncer   *realtime.StateSyncer
	sprintTaskRelationDao dao.SprintTaskRelation
}

func (s SprintTaskRelationSyncer) CreateAndSyncSprintTaskRelation(sprintTaskRelaltion entity.SprintTaskRelation, OwningTeamID uint64) error {
	err := s.sprintTaskRelationDao.CreateSprintTaskRelation(sprintTaskRelaltion)
	if err != nil {
		return err
	}

	s.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.SprintTaskRelationCollectionType,
		MutationType:   realtime.CreateMutationType,
		TeamIDs: []uint64{
			OwningTeamID,
		},
		Payload: sprintTaskRelaltion,
	})
	return nil
}

func (s SprintTaskRelationSyncer) DeleteAndSyncSprintTaskRelation(sprintID uint64, taskID uint64, OwningTeamID uint64) error {
	err := s.sprintTaskRelationDao.DeleteSprintTaskRelation(sprintID, taskID)
	if err != nil {
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

func NewSprintTaskRelationSyncer(realTimeStateSyncer *realtime.StateSyncer, sprintTaskRelationDao dao.SprintTaskRelation) SprintTaskRelationSyncer {
	return SprintTaskRelationSyncer{
		realTimeStateSyncer:   realTimeStateSyncer,
		sprintTaskRelationDao: sprintTaskRelationDao,
	}
}
