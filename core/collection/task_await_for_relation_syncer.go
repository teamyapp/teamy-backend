package collection

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type TaskAwaitForRelationSyncer struct {
	dataCollector           obs.DataCollector
	realTimeStateSyncer     *realtime.StateSyncer
	taskAwaitForRelationDao dao.TaskAwaitForRelation
	taskDao                 dao.Task
}

func (t TaskAwaitForRelationSyncer) CreateAndSyncRelation(
	ct context.Context,
	tx realtime.Transaction,
	relation entity.TaskAwaitForRelation) error {
	err := t.taskAwaitForRelationDao.CreateRelation(ct, relation)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.TaskAwaitForRelationCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        relation,
	})
	return nil
}

func (t TaskAwaitForRelationSyncer) DeleteAndSyncRelation(
	ct context.Context,
	tx realtime.Transaction,
	awaitingTaskID uint64,
	awaitForTaskID uint64) error {
	err := t.taskAwaitForRelationDao.DeleteRelation(ct, awaitingTaskID, awaitForTaskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.TaskAwaitForRelationCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload: struct {
			AwaitingTaskID uint64
			AwaitForTaskID uint64
		}{
			AwaitingTaskID: awaitingTaskID,
			AwaitForTaskID: awaitForTaskID,
		},
	})
	return nil
}

func NewTaskAwaitForRelationSyncer(
	dataCollector obs.DataCollector,
	realTimeStateSyncer *realtime.StateSyncer,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	taskDao dao.Task,
) TaskAwaitForRelationSyncer {
	return TaskAwaitForRelationSyncer{
		dataCollector:           dataCollector,
		realTimeStateSyncer:     realTimeStateSyncer,
		taskAwaitForRelationDao: taskAwaitForRelationDao,
		taskDao:                 taskDao,
	}
}
