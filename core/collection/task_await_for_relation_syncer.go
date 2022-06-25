package collection

import (
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type TaskAwaitForRelationSyncer struct {
	realTimeStateSyncer     *realtime.StateSyncer
	taskAwaitForRelationDao dao.TaskAwaitForRelation
	taskDao                 dao.Task
}

func (t TaskAwaitForRelationSyncer) CreateAndSyncRelation(relation entity.TaskAwaitForRelation) error {
	err := t.taskAwaitForRelationDao.CreateRelation(relation)
	if err != nil {
		return err
	}

	task, err := t.taskDao.FindTaskByID(relation.AwaitingTaskID)
	if err != nil {
		return err
	}

	t.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.TaskAwaitForRelationCollectionType,
		MutationType:   realtime.CreateMutationType,
		TeamIDs: []uint64{
			task.OwningTeamID,
		},
		Payload: relation,
	})
	return nil
}

func (t TaskAwaitForRelationSyncer) DeleteAndSyncRelation(awaitingTaskID uint64, awaitForTaskID uint64) error {
	err := t.taskAwaitForRelationDao.DeleteRelation(awaitingTaskID, awaitForTaskID)
	if err != nil {
		return err
	}

	task, err := t.taskDao.FindTaskByID(awaitingTaskID)
	if err != nil {
		return err
	}

	t.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.TaskAwaitForRelationCollectionType,
		MutationType:   realtime.DeleteMutationType,
		TeamIDs: []uint64{
			task.OwningTeamID,
		},
		Payload: realtime.DeleteTaskAwaitForRelationPayload{
			AwaitingTaskID: awaitingTaskID,
			AwaitForTaskID: awaitForTaskID,
		},
	})
	return nil
}

func NewTaskAwaitForRelationSyncer(
	realTimeStateSyncer *realtime.StateSyncer,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	taskDao dao.Task,
) TaskAwaitForRelationSyncer {
	return TaskAwaitForRelationSyncer{
		realTimeStateSyncer:     realTimeStateSyncer,
		taskAwaitForRelationDao: taskAwaitForRelationDao,
		taskDao:                 taskDao,
	}
}
