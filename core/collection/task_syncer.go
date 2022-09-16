package collection

import (
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type TaskSyncer struct {
	dataCollector       obs.DataCollector
	realTimeStateSyncer *realtime.StateSyncer
	taskDao             dao.Task
}

func (t TaskSyncer) CreateAndSyncTask(task entity.Task) error {
	err := t.taskDao.CreateTask(task)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	t.realTimeStateSyncer.NotifyMutation(entity.MessageEvent{
		Type: entity.MutationMessageType,
		Payload: entity.MutationPayload{
			CollectionType: entity.TaskCollectionType,
			MutationType:   entity.CreateMutationType,
			TeamIDs: []uint64{
				task.OwningTeamID,
			},
			Payload: task,
		}},
	)
	return nil
}

func (t TaskSyncer) UpdateAndSyncTask(task entity.Task) error {
	err := t.taskDao.UpdateTask(task)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	t.realTimeStateSyncer.NotifyMutation(entity.MessageEvent{
		Type: entity.MutationMessageType,
		Payload: entity.MutationPayload{
			CollectionType: entity.TaskCollectionType,
			MutationType:   entity.UpdateMutationType,
			TeamIDs: []uint64{
				task.OwningTeamID,
			},
			Payload: task},
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

	t.realTimeStateSyncer.NotifyMutation(entity.MessageEvent{
		Type: entity.MutationMessageType,
		Payload: entity.MutationPayload{
			CollectionType: entity.TaskCollectionType,
			MutationType:   entity.DeleteMutationType,
			TeamIDs: []uint64{
				task.OwningTeamID,
			},
			Payload: taskID,
		},
	})
	return nil
}

func (t TaskSyncer) StartDraggingSyncTask(taskID uint64, userID uint64, clientID uint64, owningTeamID uint64) error {
	t.realTimeStateSyncer.NotifyMutation(entity.MessageEvent{
		Type: entity.MutationMessageType,
		Payload: entity.MutationPayload{
			CollectionType: entity.TaskActivityCollectionType,
			MutationType:   entity.UpdateMutationType,
			TeamIDs: []uint64{
				owningTeamID,
			},
			Payload: entity.TeamTaskDraggingActivity{
				TaskID:           taskID,
				TeamID:           owningTeamID,
				IsDragging:       true,
				DragByUserID:     userID,
				DraggingClientID: clientID,
			}},
	})

	return nil
}

func (t TaskSyncer) StopDraggingSyncTask(taskID uint64, userID uint64, clientID uint64, owningTeamID uint64) error {
	t.realTimeStateSyncer.NotifyMutation(entity.MessageEvent{
		Type: entity.MutationMessageType,
		Payload: entity.MutationPayload{
			CollectionType: entity.TaskActivityCollectionType,
			MutationType:   entity.UpdateMutationType,
			TeamIDs: []uint64{
				owningTeamID,
			},
			Payload: entity.TeamTaskDraggingActivity{
				TaskID:           taskID,
				TeamID:           owningTeamID,
				IsDragging:       false,
				DragByUserID:     userID,
				DraggingClientID: clientID,
			}},
	})

	return nil
}

func NewTaskSyncer(
	dataCollector obs.DataCollector,
	realTimeStateSyncer *realtime.StateSyncer,
	taskDao dao.Task) TaskSyncer {
	return TaskSyncer{
		dataCollector:       dataCollector,
		realTimeStateSyncer: realTimeStateSyncer,
		taskDao:             taskDao,
	}
}
