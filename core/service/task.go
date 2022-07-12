package service

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/teamy-backend/core/collection"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

var awaitableTaskStatuses = map[entity.TaskStatus]bool{
	entity.TaskStatusInProgress: true,
	entity.TaskStatusAwaiting:   true,
}

type TaskFilter struct {
	TaskID       *uint64
	OwnerID      *uint64
	GoalContains *string
	Status       *entity.TaskStatus
}

type CreateTaskInput struct {
	Goal        string
	Context     *string
	OwnerUserID *uint64
	DueAt       *time.Time
}

type UpdateTaskInput struct {
	Goal         string
	Context      *string
	OwnerUserID  *uint64
	OwningTeamID uint64
	Effort       *int
	DueAt        *time.Time
}

type Task struct {
	cloudClientRegistry        *cloudAPI.ClientRegistry
	taskDao                    dao.Task
	threadDao                  dao.Thread
	taskAwaitForRelationDao    dao.TaskAwaitForRelation
	taskSyncer                 collection.TaskSyncer
	taskAwaitForRelationSyncer collection.TaskAwaitForRelationSyncer
	threadService              Thread
}

func (t Task) FindAllTasks(ct context.Context, filter *TaskFilter) ([]entity.Task, error) {
	tasks, err := t.taskDao.FindAllTasks()
	if err != nil {
		return nil, err
	}

	if filter != nil {
		tasks = collect.Filter(tasks, func(task entity.Task) bool {
			return matchTask(*filter, task)
		})
	}

	return tasks, nil
}

func (t Task) FindAwaitForTasks(ct context.Context, awaitingTaskID uint64) ([]entity.Task, error) {
	awaitForTaskIds, err := t.taskAwaitForRelationDao.FindAwaitForTaskIDs(awaitingTaskID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tasks, err := t.taskDao.FindTasksByIDs(awaitForTaskIds)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return tasks, nil
}

func (t Task) FindTask(ct context.Context, taskID uint64) (entity.Task, error) {
	return t.taskDao.FindTaskByID(taskID)
}

func (t Task) FindTasksInTeam(ct context.Context, teamID uint64, filter *TaskFilter) ([]entity.Task, error) {
	tasks, err := t.taskDao.FindTasksByTeamID(teamID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if filter != nil {
		tasks = collect.Filter(tasks, func(task entity.Task) bool {
			return matchTask(*filter, task)
		})
	}

	return tasks, nil
}

func (t Task) CreateTask(ct context.Context, teamID uint64, taskInput CreateTaskInput) (entity.Task, error) {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		return entity.Task{}, err
	}

	genTaskIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "taskID"}
	genTaskIDRes, err := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTaskIDReq)
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}

	threadID, err := t.threadService.createThread(ct)
	if err != nil {
		return entity.Task{}, err
	}

	task := entity.Task{
		ID:               genTaskIDRes.UniqueNumber,
		Goal:             taskInput.Goal,
		Context:          taskInput.Context,
		Status:           entity.TaskStatusUpcoming,
		CreatorUserID:    userID,
		OwningTeamID:     teamID,
		OwnerUserID:      taskInput.OwnerUserID,
		CommentsThreadID: threadID,
		CreatedAt:        time.Now(),
		DueAt:            taskInput.DueAt,
	}

	err = t.taskSyncer.CreateAndSyncTask(task)
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) UpdateTask(ct context.Context, taskID uint64, input UpdateTaskInput) (entity.Task, error) {
	task, err := t.taskDao.FindTaskByID(taskID)
	if err != nil {
		return entity.Task{}, err
	}

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
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) DeleteTask(ct context.Context, taskID uint64) (entity.Task, error) {
	task, err := t.taskDao.FindTaskByID(taskID)
	if err != nil {
		return entity.Task{}, err
	}

	err = t.taskSyncer.DeleteAndSyncTask(taskID)
	if err != nil {
		return entity.Task{}, err
	}

	err = t.threadDao.DeleteThread(task.CommentsThreadID)
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) MoveTaskToUpcoming(ct context.Context, taskID uint64) (entity.Task, error) {
	task, err := t.taskDao.FindTaskByID(taskID)
	if err != nil {
		return entity.Task{}, err
	}

	task.Status = entity.TaskStatusUpcoming
	now := time.Now()
	task.UpdatedAt = &now
	err = t.taskSyncer.UpdateAndSyncTask(task)
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) MoveTaskToInProgress(ct context.Context, taskID uint64) (entity.Task, error) {
	userID, err := ctx.UserIDFromContext(ct)
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}

	task, err := t.taskDao.FindTaskByID(taskID)
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}

	tasks, err := t.taskDao.FindTasksByTeamID(task.OwningTeamID)
	if err != nil {
		log.Println(err)
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
			log.Println(err)
			return entity.Task{}, err
		}
	}

	task.Status = entity.TaskStatusInProgress
	task.UpdatedAt = &now
	err = t.taskSyncer.UpdateAndSyncTask(task)
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) MoveTaskToDelivered(ct context.Context, taskID uint64) (entity.Task, error) {
	task, err := t.taskDao.FindTaskByID(taskID)
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}

	task.Status = entity.TaskStatusDelivered
	now := time.Now()
	task.UpdatedAt = &now
	err = t.taskSyncer.UpdateAndSyncTask(task)
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}

	awaitingTaskIDs, err := t.taskAwaitForRelationDao.FindAwaitingTaskIDs(taskID)
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}

	for _, awaitingTaskID := range awaitingTaskIDs {
		awaitForTaskIDs, err := t.taskAwaitForRelationDao.FindAwaitForTaskIDs(awaitingTaskID)
		if err != nil {
			log.Println(err)
			return entity.Task{}, err
		}

		awaitForTasks, err := t.taskDao.FindTasksByIDs(awaitForTaskIDs)
		if err != nil {
			log.Println(err)
			return entity.Task{}, err
		}

		awaitForTasks = collect.Filter(awaitForTasks, func(awaitForTask entity.Task) bool {
			return awaitForTask.Status != entity.TaskStatusDelivered
		})
		if len(awaitForTasks) == 0 {
			_, err = t.MoveTaskToUpcoming(ct, awaitingTaskID)
			if err != nil {
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
		log.Println(err)
		return entity.Task{}, err
	}

	if !awaitableTaskStatuses[task.Status] {
		return entity.Task{}, fmt.Errorf("task must be awaitable: taskID=%d", awaitingTaskID)
	}

	now := time.Now()
	err = t.taskAwaitForRelationSyncer.CreateAndSyncRelation(entity.TaskAwaitForRelation{
		AwaitingTaskID: awaitingTaskID,
		AwaitForTaskID: awaitForTaskId,
		CreatedAt:      now,
	})
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}

	task.Status = entity.TaskStatusAwaiting
	task.UpdatedAt = &now
	err = t.taskSyncer.UpdateAndSyncTask(task)
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) RemoveAwaitForTask(ct context.Context, taskID uint64, awaitForTaskId uint64) (entity.Task, error) {
	task, err := t.taskDao.FindTaskByID(taskID)
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}

	if task.Status != entity.TaskStatusAwaiting {
		return entity.Task{}, fmt.Errorf("task must be awaiting: taskID=%d", taskID)
	}

	err = t.taskAwaitForRelationSyncer.DeleteAndSyncRelation(taskID, awaitForTaskId)
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}

	awaitForTaskIds, err := t.taskAwaitForRelationDao.FindAwaitForTaskIDs(taskID)
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}

	if len(awaitForTaskIds) == 0 {
		task, err = t.MoveTaskToUpcoming(ct, taskID)
		if err != nil {
			log.Println(err)
			return entity.Task{}, err
		}
	}

	return task, nil
}

func NewTask(
	cloudClientRegistry *cloudAPI.ClientRegistry,
	taskDao dao.Task,
	threadDao dao.Thread,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	taskSyncer collection.TaskSyncer,
	taskAwaitForRelationSyncer collection.TaskAwaitForRelationSyncer,
	threadService Thread,
) Task {
	return Task{
		cloudClientRegistry:        cloudClientRegistry,
		taskDao:                    taskDao,
		threadDao:                  threadDao,
		taskAwaitForRelationDao:    taskAwaitForRelationDao,
		taskSyncer:                 taskSyncer,
		taskAwaitForRelationSyncer: taskAwaitForRelationSyncer,
		threadService:              threadService,
	}
}

func matchTask(filter TaskFilter, task entity.Task) bool {
	if filter.TaskID != nil && *filter.TaskID != task.ID {
		return false
	}

	if filter.OwnerID != nil {
		if task.OwnerUserID == nil || *filter.OwnerID != *task.OwnerUserID {
			return false
		}
	}

	if filter.Status != nil && *filter.Status != task.Status {
		return false
	}

	if filter.GoalContains != nil &&
		!strings.Contains(strings.ToLower(task.Goal), strings.ToLower(*filter.GoalContains)) {
		return false
	}

	return true
}
