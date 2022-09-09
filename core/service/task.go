package service

import (
	"context"
	"errors"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/collection"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

var awaitableTaskStatuses = map[entity.TaskStatus]bool{
	entity.TaskStatusInProgress: true,
	entity.TaskStatusAwaiting:   true,
}

type CreateTaskInput struct {
	Goal        string
	Context     *string
	OwnerUserID *uint64
	DueAt       *time.Time
	IsPlanned   *bool
}

type UpdateTaskInput struct {
	Goal         string
	Context      *string
	OwnerUserID  *uint64
	OwningTeamID uint64
	Effort       *time.Duration
	DueAt        *time.Time
}

type Task struct {
	dataCollector              obs.DataCollector
	cloudClientRegistry        *cloudAPI.ClientRegistry
	taskDao                    dao.Task
	threadDao                  dao.Thread
	taskAwaitForRelationDao    dao.TaskAwaitForRelation
	sprintParticipantDao       dao.SprintParticipant
	sprintTaskRelationDao      dao.SprintTaskRelation
	taskSyncer                 collection.TaskSyncer
	taskAwaitForRelationSyncer collection.TaskAwaitForRelationSyncer
	threadService              Thread
}

func (t Task) FindTasks(ct context.Context, filter *TaskFilter) ([]entity.Task, error) {
	tasks, err := t.taskDao.FindAllTasks()
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	if filter != nil {
		tasks = filterTasks(tasks, *filter)
	}

	return tasks, nil
}

func (t Task) FindAwaitForTasks(ct context.Context, awaitingTaskID uint64) ([]entity.Task, error) {
	awaitForTaskIds, err := t.taskAwaitForRelationDao.FindAwaitForTaskIDs(awaitingTaskID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	tasks, err := t.taskDao.FindTasksByIDs(awaitForTaskIds)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return tasks, nil
}

func (t Task) CreateTask(ct context.Context, teamID uint64, taskInput CreateTaskInput) (entity.Task, error) {
	userID, err := ctx.UserIDFromContext(t.dataCollector, ct)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	genTaskIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "taskID"}
	genTaskIDRes, err := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTaskIDReq)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	threadID, err := t.threadService.createThread(ct)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	var isPlanned bool
	if taskInput.IsPlanned != nil {
		isPlanned = *taskInput.IsPlanned
	}

	task := entity.Task{
		ID:               genTaskIDRes.UniqueNumber,
		Goal:             taskInput.Goal,
		Context:          taskInput.Context,
		Status:           entity.TaskStatusTodo,
		IsPlanned:        isPlanned,
		CreatorUserID:    userID,
		OwningTeamID:     teamID,
		OwnerUserID:      taskInput.OwnerUserID,
		CommentsThreadID: threadID,
		CreatedAt:        time.Now(),
		DueAt:            taskInput.DueAt,
	}

	err = t.taskSyncer.CreateAndSyncTask(task)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) UpdateTask(ct context.Context, taskID uint64, input UpdateTaskInput) (entity.Task, error) {
	task, err := t.taskDao.FindTaskByID(taskID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	oldEffort := task.Effort
	oldOwnerID := task.OwnerUserID
	task.Goal = input.Goal
	task.Context = input.Context
	task.OwnerUserID = input.OwnerUserID
	task.OwningTeamID = input.OwningTeamID
	task.Effort = input.Effort
	task.DueAt = input.DueAt
	updatedAt := time.Now()
	task.UpdatedAt = &updatedAt
	err = t.taskSyncer.UpdateAndSyncTask(task)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = t.updateUnusedBandWidth(taskID, oldEffort, input.Effort, oldOwnerID, input.OwnerUserID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) updateUnusedBandWidth(
	taskID uint64,
	oldEffort *time.Duration,
	newEffort *time.Duration,
	oldOwnerID *uint64,
	newOwnerID *uint64,
) error {
	err := t.tryIncreaseBandwidth(taskID, oldOwnerID, oldEffort)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	if newOwnerID != nil && newEffort != nil {
		participants, err := t.findTaskOwnerInSprints(taskID, *newOwnerID)
		if err != nil {
			t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return err
		}

		for _, participant := range participants {
			participant.UnusedBandwidth -= *newEffort
			err = t.sprintParticipantDao.UpdateSprintParticipant(participant)
			if err != nil {
				t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
				return err
			}
		}
	}

	return nil
}

func (t Task) tryIncreaseBandwidth(taskID uint64, oldOwnerID *uint64, oldEffort *time.Duration) error {
	if oldOwnerID != nil && oldEffort != nil {
		participants, err := t.findTaskOwnerInSprints(taskID, *oldOwnerID)
		if err != nil {
			t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return err
		}

		for _, participant := range participants {
			participant.UnusedBandwidth += *oldEffort
			err = t.sprintParticipantDao.UpdateSprintParticipant(participant)
			if err != nil {
				t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
				return err
			}
		}
	}

	return nil
}

func (t Task) findTaskOwnerInSprints(taskID uint64, taskOwnerUserID uint64) ([]entity.SprintParticipant, error) {
	sprintIDs, err := t.sprintTaskRelationDao.FindSprintIDsByTaskID(taskID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	participants := make([]entity.SprintParticipant, 0)
	for _, sprintID := range sprintIDs {
		participant, err := t.sprintParticipantDao.FindParticipant(sprintID, taskOwnerUserID)
		if err != nil {
			if !errors.As(err, &dao.ErrorNotFound) {
				t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
				return nil, err
			}

			continue
		}

		participants = append(participants, participant)
	}

	return participants, nil
}

func (t Task) DeleteTask(ct context.Context, taskID uint64) (entity.Task, error) {
	task, err := t.taskDao.FindTaskByID(taskID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = t.tryIncreaseBandwidth(taskID, task.OwnerUserID, task.Effort)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	// TODO: delete awaiting, await for and sprint task relationships
	err = t.taskSyncer.DeleteAndSyncTask(taskID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = t.threadDao.DeleteThread(task.CommentsThreadID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) MoveTaskToUpcoming(ct context.Context, taskID uint64, autoPauseTask bool) (entity.Task, error) {
	task, err := t.taskDao.FindTaskByID(taskID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	if autoPauseTask {
		switch task.Status {
		case entity.TaskStatusInProgress, entity.TaskStatusPaused:
			task.Status = entity.TaskStatusPaused
		}
	} else {
		task.Status = entity.TaskStatusTodo
	}

	now := time.Now()
	task.UpdatedAt = &now
	err = t.taskSyncer.UpdateAndSyncTask(task)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) MoveTaskToInProgress(ct context.Context, taskID uint64) (entity.Task, error) {
	userID, err := ctx.UserIDFromContext(t.dataCollector, ct)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	task, err := t.taskDao.FindTaskByID(taskID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	tasks, err := t.taskDao.FindTasksByTeamID(task.OwningTeamID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	if task.OwnerUserID == nil {
		task.OwnerUserID = &userID
	}

	inProgressTasks := collect.Filter(tasks, func(eachTask entity.Task) bool {
		if eachTask.OwnerUserID == nil {
			return false
		}

		if *eachTask.OwnerUserID != *task.OwnerUserID {
			return false
		}

		return eachTask.Status == entity.TaskStatusInProgress
	})

	now := time.Now()
	if len(inProgressTasks) > 0 {
		inProgressTask := inProgressTasks[0]
		inProgressTask.Status = entity.TaskStatusPaused
		inProgressTask.UpdatedAt = &now
		err = t.taskSyncer.UpdateAndSyncTask(inProgressTask)
		if err != nil {
			t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Task{}, err
		}
	}

	task.Status = entity.TaskStatusInProgress
	task.UpdatedAt = &now
	err = t.taskSyncer.UpdateAndSyncTask(task)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) MoveTaskToDelivered(ct context.Context, taskID uint64) (entity.Task, error) {
	task, err := t.taskDao.FindTaskByID(taskID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	task.Status = entity.TaskStatusDelivered
	now := time.Now()
	task.UpdatedAt = &now
	task.DeliveredAt = &now
	err = t.taskSyncer.UpdateAndSyncTask(task)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	awaitingTaskIDs, err := t.taskAwaitForRelationDao.FindAwaitingTaskIDs(taskID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	for _, awaitingTaskID := range awaitingTaskIDs {
		awaitForTaskIDs, err := t.taskAwaitForRelationDao.FindAwaitForTaskIDs(awaitingTaskID)
		if err != nil {
			t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Task{}, err
		}

		awaitForTasks, err := t.taskDao.FindTasksByIDs(awaitForTaskIDs)
		if err != nil {
			t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Task{}, err
		}

		awaitForTasks = collect.Filter(awaitForTasks, func(awaitForTask entity.Task) bool {
			return awaitForTask.Status != entity.TaskStatusDelivered
		})
		if len(awaitForTasks) == 0 {
			_, err = t.MoveTaskToUpcoming(ct, awaitingTaskID, false)
			if err != nil {
				t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
				return entity.Task{}, err
			}
		}
	}

	return task, nil
}

func (t Task) MoveTaskToBlocked(ct context.Context, taskID uint64, reason string) (entity.Task, error) {
	panic("implement me")
}

func (t Task) AddAwaitForTask(ct context.Context, awaitingTaskID uint64, awaitForTaskId uint64) (entity.Task, error) {
	task, err := t.taskDao.FindTaskByID(awaitingTaskID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	if !awaitableTaskStatuses[task.Status] {
		err = errors.New("task must be awaitable")
		t.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"taskID": awaitingTaskID,
			},
		})
		return entity.Task{}, err
	}

	now := time.Now()
	err = t.taskAwaitForRelationSyncer.CreateAndSyncRelation(entity.TaskAwaitForRelation{
		AwaitingTaskID: awaitingTaskID,
		AwaitForTaskID: awaitForTaskId,
		CreatedAt:      now,
	})
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	task.Status = entity.TaskStatusAwaiting
	task.UpdatedAt = &now
	err = t.taskSyncer.UpdateAndSyncTask(task)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) RemoveAwaitForTask(ct context.Context, taskID uint64, awaitForTaskId uint64) (entity.Task, error) {
	task, err := t.taskDao.FindTaskByID(taskID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	if task.Status != entity.TaskStatusAwaiting {
		err = errors.New("task must be awaitable")
		t.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"taskID": taskID,
			},
		})
		return entity.Task{}, err
	}

	err = t.taskAwaitForRelationSyncer.DeleteAndSyncRelation(taskID, awaitForTaskId)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	awaitForTaskIds, err := t.taskAwaitForRelationDao.FindAwaitForTaskIDs(taskID)
	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	if len(awaitForTaskIds) == 0 {
		task, err = t.MoveTaskToUpcoming(ct, taskID, false)
		if err != nil {
			t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Task{}, err
		}
	}

	return task, nil
}

func NewTask(
	dataCollector obs.DataCollector,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	taskDao dao.Task,
	threadDao dao.Thread,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	sprintParticipantDao dao.SprintParticipant,
	sprintTaskRelationDao dao.SprintTaskRelation,
	taskSyncer collection.TaskSyncer,
	taskAwaitForRelationSyncer collection.TaskAwaitForRelationSyncer,
	threadService Thread,
) Task {
	return Task{
		dataCollector:              dataCollector,
		cloudClientRegistry:        cloudClientRegistry,
		taskDao:                    taskDao,
		threadDao:                  threadDao,
		taskAwaitForRelationDao:    taskAwaitForRelationDao,
		sprintParticipantDao:       sprintParticipantDao,
		sprintTaskRelationDao:      sprintTaskRelationDao,
		taskSyncer:                 taskSyncer,
		taskAwaitForRelationSyncer: taskAwaitForRelationSyncer,
		threadService:              threadService,
	}
}
