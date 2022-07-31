package collection

import (
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type TaskSyncer struct {
	realTimeStateSyncer *realtime.StateSyncer
	taskDao             dao.Task
}

func (t TaskSyncer) CreateAndSyncTask(task entity.Task) error {
	err := t.taskDao.CreateTask(task)
	if err != nil {
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
		return err
	}

	err = t.taskDao.DeleteTask(taskID)
	if err != nil {
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

func NewTaskSyncer(realTimeStateSyncer *realtime.StateSyncer, taskDao dao.Task) TaskSyncer {
	return TaskSyncer{
		realTimeStateSyncer: realTimeStateSyncer,
		taskDao:             taskDao,
	}
}
