package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/collection"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

var awaitableTaskStatuses = map[entity.TaskStatus]bool{
	entity.TaskStatusInProgress: true,
	entity.TaskStatusAwaiting:   true,
}

type CreateTaskInput struct {
	Goal        string
	Context     *string
	OwnerUserID *uint64
	IsPlanned   *bool
	DueAt       *time.Time
}

type createTaskInput struct {
	Goal          string
	DueAt         *time.Time
	Context       *string
	CreatorUserID uint64
	OwnerUserID   *uint64
	Status        entity.TaskStatus
	IsPlanned     bool
	Effort        *time.Duration
	UpdatedAt     *time.Time
	DeliveredAt   *time.Time
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
	authorizer                 Authorizer
	activityCache              cache.Activity
	taskDao                    dao.Task
	threadDao                  dao.Thread
	taskAwaitForRelationDao    dao.TaskAwaitForRelation
	sprintParticipantDao       dao.SprintParticipant
	sprintTaskRelationDao      dao.SprintTaskRelation
	taskSyncer                 collection.TaskSyncer
	taskAwaitForRelationSyncer collection.TaskAwaitForRelationSyncer
	sprintParticipantSyncer    collection.SprintParticipantSyncer
	threadService              Thread
	stateSyncer                *realtime.StateSyncer
}

func (t Task) FindTaskByID(ct context.Context, taskID uint64) (entity.Task, error) {
	return t.taskDao.FindTaskByID(ct, taskID)
}

func (t Task) FindTasks(ct context.Context, filter *TaskFilter) ([]entity.Task, error) {
	tasks, err := t.taskDao.FindAllTasks(ct)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	if filter != nil {
		tasks = filterTasks(tasks, *filter)
	}

	return tasks, nil
}

func (t Task) FindAwaitForTasks(ct context.Context, awaitingTaskID uint64) ([]entity.Task, error) {
	awaitForTaskIds, err := t.taskAwaitForRelationDao.FindAwaitForTaskIDs(ct, awaitingTaskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	tasks, err := t.taskDao.FindTasksByIDs(ct, awaitForTaskIds)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return tasks, nil
}

func (t Task) createTask(ct context.Context, teamID uint64, taskInput createTaskInput) (entity.Task, error) {
	transaction := realtime.NewTransaction(t.stateSyncer, t.dataCollector, teamID)
	genTaskIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "taskID"}
	genTaskIDRes, err := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTaskIDReq)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	threadID, err := t.threadService.createThread(ct)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	task := entity.Task{
		ID:               genTaskIDRes.UniqueNumber,
		Goal:             taskInput.Goal,
		Context:          taskInput.Context,
		Status:           taskInput.Status,
		IsPlanned:        taskInput.IsPlanned,
		CreatorUserID:    taskInput.CreatorUserID,
		OwningTeamID:     teamID,
		Effort:           taskInput.Effort,
		OwnerUserID:      taskInput.OwnerUserID,
		CommentsThreadID: threadID,
		CreatedAt:        time.Now(),
		DueAt:            taskInput.DueAt,
		DeliveredAt:      taskInput.DeliveredAt,
	}

	err = t.taskSyncer.CreateAndSyncTask(ct, *transaction, task)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = t.stateSyncer.ProcessTransaction(ct, transaction)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	if feature.EnableAuthorization {
		err = t.authorizer.registerResource(ct, authorization.TaskResourceType, task.ID)
		if err != nil {
			t.authorizer.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Task{}, err
		}

		err = t.authorizer.assignParentResource(ct, authorization.TaskResourceType, task.ID, authorization.TeamResourceType, teamID)
		if err != nil {
			t.authorizer.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Task{}, err
		}
	}

	return task, nil
}

func (t Task) CreateTask(ct context.Context, teamID uint64, taskInput CreateTaskInput) (entity.Task, error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	if feature.EnableAuthorization {
		query := authorization.NewCreateTaskQuery(userID, teamID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			t.authorizer.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Task{}, err
		}

		if !hasPermission {
			return entity.Task{}, authorization.Error{
				Code:    authorization.UnauthorizedErrorCode,
				Message: fmt.Sprintf("Unauthorized: %v", query),
			}
		}
	}
	var isPlanned bool
	if taskInput.IsPlanned != nil {
		isPlanned = *taskInput.IsPlanned
	}

	input := createTaskInput{
		IsPlanned:     isPlanned,
		Goal:          taskInput.Goal,
		Context:       taskInput.Context,
		Status:        entity.TaskStatusTodo,
		CreatorUserID: userID,
		OwnerUserID:   taskInput.OwnerUserID,
	}

	return t.createTask(ct, teamID, input)
}

func (t Task) UpdateTask(ct context.Context, taskID uint64, input UpdateTaskInput) (entity.Task, error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		t.authorizer.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	if feature.EnableAuthorization {
		query := authorization.NewUpdateTaskQuery(userID, taskID)
		hasPermission, err := t.authorizer.hasPermission(ct, query)
		if err != nil {
			t.authorizer.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Task{}, err
		}

		if !hasPermission {
			return entity.Task{}, authorization.Error{
				Code:    authorization.UnauthorizedErrorCode,
				Message: fmt.Sprintf("Unauthorized: %v", query),
			}
		}
	}

	task, err := t.taskDao.FindTaskByID(ct, taskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	transaction := realtime.NewTransaction(t.stateSyncer, t.dataCollector, task.OwningTeamID)
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
	err = t.taskSyncer.UpdateAndSyncTask(ct, *transaction, task)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = t.updateUnusedBandWidth(ct, *transaction, taskID, oldEffort, input.Effort, oldOwnerID, input.OwnerUserID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = t.stateSyncer.ProcessTransaction(ct, transaction)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}
	return task, nil
}

func (t Task) updateUnusedBandWidth(
	ct context.Context,
	tx realtime.Transaction,
	taskID uint64,
	oldEffort *time.Duration,
	newEffort *time.Duration,
	oldOwnerID *uint64,
	newOwnerID *uint64,
) error {
	err := t.tryIncreaseBandwidth(ct, tx, taskID, oldOwnerID, oldEffort)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	if newOwnerID != nil && newEffort != nil {
		participants, err := t.findTaskOwnerInSprints(ct, taskID, *newOwnerID)
		if err != nil {
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return err
		}

		for _, participant := range participants {
			participant.UnusedBandwidth -= *newEffort
			err = t.sprintParticipantSyncer.UpdateAndSyncSprintParticipant(ct, tx, participant)
			if err != nil {
				t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
				return err
			}
		}
	}

	return nil
}

func (t Task) tryIncreaseBandwidth(ct context.Context, tx realtime.Transaction, taskID uint64, oldOwnerID *uint64, oldEffort *time.Duration) error {
	if oldOwnerID != nil && oldEffort != nil {
		participants, err := t.findTaskOwnerInSprints(ct, taskID, *oldOwnerID)
		if err != nil {
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return err
		}

		for _, participant := range participants {
			participant.UnusedBandwidth += *oldEffort
			err = t.sprintParticipantSyncer.UpdateAndSyncSprintParticipant(ct, tx, participant)
			if err != nil {
				t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
				return err
			}
		}
	}

	return nil
}

func (t Task) findTaskOwnerInSprints(ct context.Context, taskID uint64, taskOwnerUserID uint64) ([]entity.SprintParticipant, error) {
	sprintIDs, err := t.sprintTaskRelationDao.FindSprintIDsByTaskID(ct, taskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	participants := make([]entity.SprintParticipant, 0)
	for _, sprintID := range sprintIDs {
		participant, err := t.sprintParticipantDao.FindParticipant(ct, sprintID, taskOwnerUserID)
		if err != nil {
			if !errors.As(err, &dao.ErrorNotFound) {
				t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
				return nil, err
			}

			continue
		}

		participants = append(participants, participant)
	}

	return participants, nil
}

func (t Task) DeleteTask(ct context.Context, taskID uint64) (entity.Task, error) {
	task, err := t.taskDao.FindTaskByID(ct, taskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	transaction := realtime.NewTransaction(t.stateSyncer, t.dataCollector, task.OwningTeamID)
	err = t.tryIncreaseBandwidth(ct, *transaction, taskID, task.OwnerUserID, task.Effort)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	// TODO: delete awaiting, await for and sprint task relationships
	err = t.taskSyncer.DeleteAndSyncTask(ct, *transaction, taskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = t.threadDao.DeleteThread(ct, task.CommentsThreadID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = t.stateSyncer.ProcessTransaction(ct, transaction)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) moveTaskToUpcoming(ct context.Context, tx realtime.Transaction, taskID uint64, autoPauseTask bool) (entity.Task, error) {
	task, err := t.taskDao.FindTaskByID(ct, taskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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

	err = t.taskSyncer.UpdateAndSyncTask(ct, tx, task)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	return task, err
}

func (t Task) MoveTaskToUpcoming(ct context.Context, taskID uint64, autoPauseTask bool) (entity.Task, error) {
	task, err := t.taskDao.FindTaskByID(ct, taskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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

	transaction := realtime.NewTransaction(t.stateSyncer, t.dataCollector, task.OwningTeamID)
	err = t.taskSyncer.UpdateAndSyncTask(ct, *transaction, task)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = t.stateSyncer.ProcessTransaction(ct, transaction)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) MoveTaskToInProgress(ct context.Context, taskID uint64) (entity.Task, error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	task, err := t.taskDao.FindTaskByID(ct, taskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	tasks, err := t.taskDao.FindTasksByTeamID(ct, task.OwningTeamID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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

	transaction := realtime.NewTransaction(t.stateSyncer, t.dataCollector, task.OwningTeamID)

	now := time.Now()
	if len(inProgressTasks) > 0 {
		inProgressTask := inProgressTasks[0]
		inProgressTask.Status = entity.TaskStatusPaused
		inProgressTask.UpdatedAt = &now
		err = t.taskSyncer.UpdateAndSyncTask(ct, *transaction, inProgressTask)
		if err != nil {
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Task{}, err
		}
	}

	task.Status = entity.TaskStatusInProgress
	task.UpdatedAt = &now
	err = t.taskSyncer.UpdateAndSyncTask(ct, *transaction, task)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = t.stateSyncer.ProcessTransaction(ct, transaction)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) MoveTaskToDelivered(ct context.Context, taskID uint64) (entity.Task, error) {
	task, err := t.taskDao.FindTaskByID(ct, taskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	transaction := realtime.NewTransaction(t.stateSyncer, t.dataCollector, task.OwningTeamID)
	task.Status = entity.TaskStatusDelivered
	now := time.Now()
	task.UpdatedAt = &now
	task.DeliveredAt = &now
	err = t.taskSyncer.UpdateAndSyncTask(ct, *transaction, task)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	awaitingTaskIDs, err := t.taskAwaitForRelationDao.FindAwaitingTaskIDs(ct, taskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	for _, awaitingTaskID := range awaitingTaskIDs {
		awaitForTaskIDs, err := t.taskAwaitForRelationDao.FindAwaitForTaskIDs(ct, awaitingTaskID)
		if err != nil {
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Task{}, err
		}

		awaitForTasks, err := t.taskDao.FindTasksByIDs(ct, awaitForTaskIDs)
		if err != nil {
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Task{}, err
		}

		awaitForTasks = collect.Filter(awaitForTasks, func(awaitForTask entity.Task) bool {
			return awaitForTask.Status != entity.TaskStatusDelivered
		})
		if len(awaitForTasks) == 0 {
			_, err = t.moveTaskToUpcoming(ct, *transaction, awaitingTaskID, false)
			if err != nil {
				t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
				return entity.Task{}, err
			}
		}
	}

	err = t.stateSyncer.ProcessTransaction(ct, transaction)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) MoveTaskToBlocked(ct context.Context, taskID uint64, reason string) (entity.Task, error) {
	panic("implement me")
}

func (t Task) AddAwaitForTask(ct context.Context, awaitingTaskID uint64, awaitForTaskId uint64) (entity.Task, error) {
	task, err := t.taskDao.FindTaskByID(ct, awaitingTaskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	transaction := realtime.NewTransaction(t.stateSyncer, t.dataCollector, task.OwningTeamID)
	if !awaitableTaskStatuses[task.Status] {
		err = errors.New("task must be awaitable")
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"TaskID": awaitingTaskID,
			},
		})
		return entity.Task{}, err
	}

	now := time.Now()
	err = t.taskAwaitForRelationSyncer.CreateAndSyncRelation(ct, *transaction, entity.TaskAwaitForRelation{
		AwaitingTaskID: awaitingTaskID,
		AwaitForTaskID: awaitForTaskId,
		CreatedAt:      now,
	})
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	task.Status = entity.TaskStatusAwaiting
	task.UpdatedAt = &now
	err = t.taskSyncer.UpdateAndSyncTask(ct, *transaction, task)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = t.stateSyncer.ProcessTransaction(ct, transaction)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) RemoveAwaitForTask(ct context.Context, taskID uint64, awaitForTaskId uint64) (entity.Task, error) {
	task, err := t.taskDao.FindTaskByID(ct, taskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	transaction := realtime.NewTransaction(t.stateSyncer, t.dataCollector, task.OwningTeamID)
	if task.Status != entity.TaskStatusAwaiting {
		err = errors.New("task must be awaitable")
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"TaskID": taskID,
			},
		})
		return entity.Task{}, err
	}

	err = t.taskAwaitForRelationSyncer.DeleteAndSyncRelation(ct, *transaction, taskID, awaitForTaskId)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	awaitForTaskIds, err := t.taskAwaitForRelationDao.FindAwaitForTaskIDs(ct, taskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	if len(awaitForTaskIds) == 0 {
		task, err = t.moveTaskToUpcoming(ct, *transaction, taskID, false)
		if err != nil {
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Task{}, err
		}
	}

	err = t.stateSyncer.ProcessTransaction(ct, transaction)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	return task, nil
}

func (t Task) StartDraggingTask(ct context.Context, taskID uint64, clientID uint64) error {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		err := errors.New("user id not found")
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	task, err := t.taskDao.FindTaskByID(ct, taskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	transaction := realtime.NewTransaction(t.stateSyncer, t.dataCollector, task.OwningTeamID)
	err = t.taskSyncer.UpdateAndSyncTaskActivity(
		ct,
		*transaction,
		entity.TaskActivity{
			TaskID: taskID,
			TeamID: task.OwningTeamID,
			DragTaskActivity: entity.DragTaskActivity{
				IsDragging: true,
				Client: &entity.Client{
					ID:     clientID,
					UserID: userID,
				},
			}})

	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	err = t.stateSyncer.ProcessTransaction(ct, transaction)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}
	return nil
}

func (t Task) StopDraggingTask(ct context.Context, taskID uint64, clientID uint64) error {
	task, err := t.taskDao.FindTaskByID(ct, taskID)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	transaction := realtime.NewTransaction(t.stateSyncer, t.dataCollector, task.OwningTeamID)
	err = t.taskSyncer.UpdateAndSyncTaskActivity(
		ct,
		*transaction,
		entity.TaskActivity{
			TaskID: taskID,
			TeamID: task.OwningTeamID,
			DragTaskActivity: entity.DragTaskActivity{
				IsDragging: false,
				Client:     nil,
			}})

	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}

	err = t.stateSyncer.ProcessTransaction(ct, transaction)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return err
	}
	return nil
}

func NewTask(
	dataCollector obs.DataCollector,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	authorizer Authorizer,
	activityCache cache.Activity,
	taskDao dao.Task,
	threadDao dao.Thread,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	sprintParticipantDao dao.SprintParticipant,
	sprintTaskRelationDao dao.SprintTaskRelation,
	taskSyncer collection.TaskSyncer,
	taskAwaitForRelationSyncer collection.TaskAwaitForRelationSyncer,
	sprintParticipantSyncer collection.SprintParticipantSyncer,
	threadService Thread,
	stateSyncer *realtime.StateSyncer,
) Task {
	return Task{
		dataCollector:              dataCollector,
		cloudClientRegistry:        cloudClientRegistry,
		authorizer:                 authorizer,
		activityCache:              activityCache,
		taskDao:                    taskDao,
		threadDao:                  threadDao,
		taskAwaitForRelationDao:    taskAwaitForRelationDao,
		sprintParticipantDao:       sprintParticipantDao,
		sprintTaskRelationDao:      sprintTaskRelationDao,
		taskSyncer:                 taskSyncer,
		taskAwaitForRelationSyncer: taskAwaitForRelationSyncer,
		sprintParticipantSyncer:    sprintParticipantSyncer,
		threadService:              threadService,
		stateSyncer:                stateSyncer,
	}
}
