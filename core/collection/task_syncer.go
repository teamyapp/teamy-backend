package collection

import (
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

func (t TaskSyncer) CreateAndSyncTask(task entity.Task) error {
	err := t.taskDao.CreateTask(task)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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

func (t TaskSyncer) UpdateAndSyncTask(task entity.Task) error {
	err := t.taskDao.UpdateTask(task)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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

func (t TaskSyncer) DeleteAndSyncTask(taskID uint64) error {
	task, err := t.taskDao.FindTaskByID(taskID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	err = t.taskDao.DeleteTask(taskID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
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

func (t TaskSyncer) UpdateAndSyncTaskActivity(taskID uint64, clientID uint64, taskActivity entity.TaskActivity) error {
	task, err := t.taskDao.FindTaskByID(taskID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	_, err = t.activityCache.UpdateTaskActivity(task.OwningTeamID, taskID, &taskActivity)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	t.realTimeStateSyncer.NotifyMutation(realtime.Mutation{
		CollectionType: realtime.TaskActivityCollectionType,
		MutationType:   realtime.UpdateMutationType,
		TeamIDs: []uint64{
			task.OwningTeamID,
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
