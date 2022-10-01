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

func (t TaskSyncer) CreateAndSyncTask(ct context.Context, task entity.Task) error {
	err := t.taskDao.CreateTask(ct, task)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	t.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.TaskCollectionType,
		MutationType:   realtime.CreateMutationType,
		TeamIDs: []uint64{
			task.OwningTeamID,
		},
		Payload: task,
	})
	return nil
}

func (t TaskSyncer) UpdateAndSyncTask(ct context.Context, task entity.Task) error {
	err := t.taskDao.UpdateTask(ct, task)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	t.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.TaskCollectionType,
		MutationType:   realtime.UpdateMutationType,
		TeamIDs: []uint64{
			task.OwningTeamID,
		},
		Payload: task,
	})
	return nil
}

func (t TaskSyncer) DeleteAndSyncTask(ct context.Context, taskID uint64) error {
	task, err := t.taskDao.FindTaskByID(ct, taskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	err = t.taskDao.DeleteTask(ct, taskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	t.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.TaskCollectionType,
		MutationType:   realtime.DeleteMutationType,
		TeamIDs: []uint64{
			task.OwningTeamID,
		},
		Payload: taskID,
	})
	return nil
}

func (t TaskSyncer) UpdateAndSyncTaskActivity(
	ct context.Context,
	taskActivity entity.TaskActivity,
) error {
	_, err := t.activityCache.UpdateTaskActivity(ct, taskActivity.TeamID, taskActivity.TaskID, &taskActivity)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	t.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.TaskActivityCollectionType,
		MutationType:   realtime.UpdateMutationType,
		TeamIDs: []uint64{
			taskActivity.TeamID,
		},
		Payload: taskActivity},
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
