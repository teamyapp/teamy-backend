package collection

import (
	"context"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type TaskSyncer struct {
	dataCollector       obs.DataCollector
	realTimeStateSyncer *realtime.StateSyncer
	taskDao             dao.Task
	activityCache       cache.Activity
}

func (t TaskSyncer) CreateAndSyncTask(ct context.Context, tx realtime.Transaction, task entity.Task) error {
	err := t.taskDao.CreateTask(ct, task)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.TaskCollectionType,
		MutationType:   realtime.CreateMutationType,
		Payload:        task,
	})
	return nil
}

func (t TaskSyncer) UpdateAndSyncTask(ct context.Context, tx realtime.Transaction, task entity.Task) error {
	err := t.taskDao.UpdateTask(ct, task)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.TaskCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        task,
	})
	return nil
}

func (t TaskSyncer) DeleteAndSyncTask(ct context.Context, tx realtime.Transaction, taskID uint64) error {
	err := t.taskDao.DeleteTask(ct, taskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.TaskCollectionType,
		MutationType:   realtime.DeleteMutationType,
		Payload:        taskID,
	})
	return nil
}

func (t TaskSyncer) UpdateAndSyncTaskActivity(
	ct context.Context,
	tx realtime.Transaction,
	taskActivity entity.TaskActivity,
) error {
	_, err := t.activityCache.UpdateTaskActivity(ct, taskActivity.TeamID, taskActivity.TaskID, &taskActivity)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	tx.AddMutation(ct, realtime.MutationInput{
		CollectionType: realtime.TaskActivityCollectionType,
		MutationType:   realtime.UpdateMutationType,
		Payload:        taskActivity},
	)
	return nil
}

func NewTaskSyncer(
	dataCollector obs.DataCollector,
	realTimeStateSyncer *realtime.StateSyncer,
	taskDao dao.Task,
	activityCache cache.Activity) TaskSyncer {
	return TaskSyncer{
		dataCollector:       dataCollector,
		realTimeStateSyncer: realTimeStateSyncer,
		taskDao:             taskDao,
		activityCache:       activityCache,
	}
}
