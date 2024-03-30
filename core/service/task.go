package service

import (
	"context"
	"fmt"
	"time"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	cloudAuthorization "github.com/teamyapp/cloud/libs/authorization"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/cache"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/transaction"
)

var awaitableTaskStatuses = map[entity.TaskStatus]bool{
	entity.TaskStatusInProgress: true,
	entity.TaskStatusAwaiting:   true,
}

type CreateTaskInput struct {
	Goal        string
	Context     *string
	OwnerUserID *uint64
	IsPlanned   bool
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
	Priority     *entity.Priority
	DueAt        *time.Time
}

type Task struct {
	logger                  telemetry.Logger
	cloudClientRegistry     *client.Registry
	authorizer              client.Authorizer
	featureToggles          feature.Toggles
	stateSyncer             *realtime.StateSyncer
	transactionFactory      cloudTransaction.Factory
	activityCache           cache.Activity
	taskDao                 dao.Task
	sprintDao               dao.Sprint
	threadDao               dao.Thread
	taskAwaitForRelationDao dao.TaskAwaitForRelation
	sprintParticipantDao    dao.SprintParticipant
	sprintTaskRelationDao   dao.SprintTaskRelation
	storyTaskRelationDao    dao.StoryTaskRelation
}

func (t Task) FindTaskByID(ct context.Context, taskID uint64) (entity.Task, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Task{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if t.featureToggles.EnableAuthorization {
		query := authorization.NewReadInTaskQuery(userID, taskID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Task{}, err
		}

		if !hasPermission {
			return entity.Task{}, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	return t.taskDao.FindTaskByID(ct, taskID)
}

func (t Task) FindTasks(ct context.Context, filter *TaskFilter) ([]entity.Task, *errs.Error) {
	tasks, err := t.taskDao.FindAllTasks(ct)
	if err != nil {
		return nil, err
	}

	if t.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		authorizedTasks, err := client.FilterAuthorizedItems(
			ct,
			t.authorizer,
			tasks,
			func(task entity.Task) cloudAuthorization.Query {
				return authorization.NewReadInTaskQuery(userID, task.ID)
			})
		if err != nil {
			return nil, err
		}

		tasks = authorizedTasks
	}

	if filter != nil {
		tasks = filterTasks(tasks, *filter)
	}

	return tasks, nil
}

func (t Task) FindTasksInTeam(ct context.Context, teamID uint64, filter *TaskFilter) ([]entity.Task, *errs.Error) {
	tasks, err := t.taskDao.FindTasksByTeamID(ct, teamID)
	if err != nil {
		return nil, err
	}

	if t.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		authorizedTasks, err := client.FilterAuthorizedItems(
			ct,
			t.authorizer,
			tasks,
			func(task entity.Task) cloudAuthorization.Query {
				return authorization.NewReadInTaskQuery(userID, task.ID)
			})
		if err != nil {
			return nil, err
		}

		tasks = authorizedTasks
	}

	if filter != nil {
		tasks = filterTasks(tasks, *filter)
	}

	return tasks, nil
}

func (t Task) FindTasksInSprint(
	ct context.Context,
	sprintID uint64,
	filter *TaskFilter,
) ([]entity.Task, *errs.Error) {
	var tasks []entity.Task
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		taskIDs, internalErr := t.sprintTaskRelationDao.FindTaskIDsBySprintIDWithTx(ct, tx, sprintID)
		if internalErr != nil {
			return internalErr
		}

		tasks, internalErr = t.taskDao.FindTasksByIDsWithTx(ct, tx, taskIDs)
		return internalErr
	})

	if err != nil {
		return nil, err
	}

	if t.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		authorizedTasks, err := client.FilterAuthorizedItems(
			ct,
			t.authorizer,
			tasks,
			func(task entity.Task) cloudAuthorization.Query {
				return authorization.NewReadInTaskQuery(userID, task.ID)
			})
		if err != nil {
			return nil, err
		}

		tasks = authorizedTasks
	}

	if filter != nil {
		tasks = filterTasks(tasks, *filter)
	}

	return tasks, nil
}

func (t Task) FindAwaitForTasks(ct context.Context, awaitingTaskID uint64) ([]entity.Task, *errs.Error) {
	var tasks []entity.Task
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(true, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		taskIDs, internalErr := t.taskAwaitForRelationDao.FindAwaitForTaskIDsWithTx(ct, tx, awaitingTaskID)
		if internalErr != nil {
			return internalErr
		}

		tasks, internalErr = t.taskDao.FindTasksByIDsWithTx(ct, tx, taskIDs)
		return internalErr
	})

	if err != nil {
		return nil, err
	}

	if t.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		// TODO: let users know the task cannot be delivered because of
		authorizedTasks, err := client.FilterAuthorizedItems(
			ct,
			t.authorizer,
			tasks,
			func(task entity.Task) cloudAuthorization.Query {
				return authorization.NewReadInTaskQuery(userID, task.ID)
			})
		if err != nil {
			return nil, err
		}

		tasks = authorizedTasks
	}

	return tasks, err
}

func (t Task) FindTaskActivities(ct context.Context, teamID uint64) []entity.TaskActivity {
	var taskActivities []entity.TaskActivity
	taskActivityMap := t.activityCache.FindAllTaskActivitiesByTeamID(teamID)
	for _, taskActivity := range taskActivityMap {
		taskActivities = append(taskActivities, *taskActivity)
	}

	if t.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return nil
		}

		authorizedTaskActivities, err := client.FilterAuthorizedItems(
			ct,
			t.authorizer,
			taskActivities,
			func(taskActivity entity.TaskActivity) cloudAuthorization.Query {
				return authorization.NewReadInTaskQuery(userID, taskActivity.TaskID)
			})
		if err != nil {
			return nil
		}

		taskActivities = authorizedTaskActivities
	}

	return taskActivities
}

func (t Task) createTask(ct context.Context, teamID uint64, taskInput createTaskInput) (entity.Task, *errs.Error) {
	genTaskIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "taskID"}
	genTaskIDRes, rpcErr := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTaskIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.Task{}, internalErr
	}

	genThreadIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "threadID"}
	genThreadIDRes, rpcErr := t.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genThreadIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.Task{}, internalErr
	}

	var task entity.Task
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	internalErr := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		threadID := genThreadIDRes.UniqueNumber
		internalErr := t.threadDao.CreateThread(ct, tx, threadID)
		if internalErr != nil {
			return internalErr
		}

		task = entity.Task{
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
			CreatedAt:        time.Now().UTC(),
			DueAt:            taskInput.DueAt,
			DeliveredAt:      taskInput.DeliveredAt,
		}

		createTaskMutation := mutation.NewCreateTask(
			t.logger,
			t.stateSyncer,
			t.taskDao,
			task,
		)

		internalErr = createTaskMutation.Execute(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		rtTx.AppendMutation(createTaskMutation)
		return nil
	})

	if internalErr != nil {
		return entity.Task{}, internalErr
	}

	if t.featureToggles.EnableAuthorization {
		internalErr = t.authorizer.RegisterResource(ct, authorization.TaskResourceType, task.ID)
		if internalErr != nil {
			return entity.Task{}, internalErr
		}

		internalErr = t.authorizer.AssignParentResource(ct, authorization.TaskResourceType, task.ID, authorization.TeamResourceType, teamID)
		if internalErr != nil {
			return entity.Task{}, internalErr
		}
	}

	return task, nil
}

func (t Task) CreateTask(ct context.Context, teamID uint64, taskInput CreateTaskInput) (entity.Task, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Task{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if t.featureToggles.EnableAuthorization {
		query := authorization.NewCreateTaskInTeamQuery(userID, teamID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Task{}, err
		}

		if !hasPermission {
			return entity.Task{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}
	input := createTaskInput{
		IsPlanned:     taskInput.IsPlanned,
		Goal:          taskInput.Goal,
		Context:       taskInput.Context,
		Status:        entity.TaskStatusTodo,
		CreatorUserID: userID,
		OwnerUserID:   taskInput.OwnerUserID,
		DueAt:         taskInput.DueAt,
	}

	return t.createTask(ct, teamID, input)
}

func (t Task) UpdateTask(ct context.Context, taskID uint64, input UpdateTaskInput) (entity.Task, *errs.Error) {
	if t.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.Task{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewUpdateInTaskQuery(userID, taskID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Task{}, err
		}

		if !hasPermission {
			return entity.Task{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var task entity.Task
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		task, internalErr = t.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
		if internalErr != nil {
			return internalErr
		}

		oldEffort := task.Effort
		oldOwnerID := task.OwnerUserID
		task.Goal = input.Goal
		task.Context = input.Context
		task.OwnerUserID = input.OwnerUserID
		task.OwningTeamID = input.OwningTeamID
		task.Effort = input.Effort
		task.Priority = input.Priority
		task.DueAt = input.DueAt
		updatedAt := time.Now().UTC()
		task.UpdatedAt = &updatedAt
		updateTaskMutation := mutation.NewUpdateTask(
			t.logger,
			t.stateSyncer,
			t.taskDao,
			task,
		)
		rtTx.AppendMutation(updateTaskMutation)
		internalErr = updateTaskMutation.Execute(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		internalErr = t.updateUnusedBandWidth(ct, tx, rtTx, taskID, oldEffort, input.Effort, oldOwnerID, input.OwnerUserID)
		if internalErr != nil {
			return internalErr
		}

		return nil
	})

	return task, err
}

func (t Task) updateUnusedBandWidth(
	ct context.Context,
	tx *cloudTransaction.Transaction,
	rtTx *realtime.Transaction,
	taskID uint64,
	oldEffort *time.Duration,
	newEffort *time.Duration,
	oldOwnerID *uint64,
	newOwnerID *uint64,
) *errs.Error {
	internalErr := t.tryIncreaseBandwidth(ct, tx, rtTx, taskID, oldOwnerID, oldEffort)
	if internalErr != nil {
		return internalErr
	}

	if newOwnerID != nil && newEffort != nil {
		var participants []entity.SprintParticipant
		participants, internalErr = t.findTaskOwnerInSprints(ct, tx, taskID, *newOwnerID)
		if internalErr != nil {
			return internalErr
		}

		for _, participant := range participants {
			participant.UnusedBandwidth -= *newEffort
			updateSprintParticipantMutation := mutation.NewUpdateSprintParticipant(
				t.logger,
				t.stateSyncer,
				t.sprintParticipantDao,
				t.sprintDao,
				participant,
			)
			rtTx.AppendMutation(updateSprintParticipantMutation)
			internalErr = updateSprintParticipantMutation.Execute(ct, tx)
			if internalErr != nil {
				return internalErr
			}
		}
	}

	return nil
}

func (t Task) tryIncreaseBandwidth(
	ct context.Context,
	tx *cloudTransaction.Transaction,
	rtTx *realtime.Transaction,
	taskID uint64,
	oldOwnerID *uint64,
	oldEffort *time.Duration) *errs.Error {
	if oldOwnerID != nil && oldEffort != nil {
		participants, internalErr := t.findTaskOwnerInSprints(ct, tx, taskID, *oldOwnerID)
		if internalErr != nil {
			return internalErr
		}

		for _, participant := range participants {
			participant.UnusedBandwidth += *oldEffort
			updateSprintParticipantMutation := mutation.NewUpdateSprintParticipant(
				t.logger,
				t.stateSyncer,
				t.sprintParticipantDao,
				t.sprintDao,
				participant,
			)
			rtTx.AppendMutation(updateSprintParticipantMutation)
			internalErr = updateSprintParticipantMutation.Execute(ct, tx)
			if internalErr != nil {
				return internalErr
			}
		}
	}

	return nil
}

func (t Task) findTaskOwnerInSprints(ct context.Context, tx *cloudTransaction.Transaction, taskID uint64, taskOwnerUserID uint64) ([]entity.SprintParticipant, *errs.Error) {
	sprintIDs, err := t.sprintTaskRelationDao.FindSprintIDsByTaskIDWithTx(ct, tx, taskID)
	if err != nil {
		return nil, err
	}

	participants := make([]entity.SprintParticipant, 0)
	for _, sprintID := range sprintIDs {
		participant, internalErr := t.sprintParticipantDao.FindParticipantWithTx(ct, tx, sprintID, taskOwnerUserID)
		if internalErr != nil {
			if internalErr.Code != errs.NotFound {
				return nil, internalErr
			}

			continue
		}

		participants = append(participants, participant)
	}

	return participants, nil
}

func (t Task) DeleteTask(ct context.Context, taskID uint64) (entity.Task, *errs.Error) {
	if t.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.Task{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewDeleteInTaskQuery(userID, taskID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Task{}, err
		}

		if !hasPermission {
			return entity.Task{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var task entity.Task
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var internalErr *errs.Error
		task, internalErr = t.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
		if internalErr != nil {
			return internalErr
		}

		sprintIDs, internalErr := t.sprintTaskRelationDao.FindSprintIDsByTaskIDWithTx(ct, tx, taskID)
		if internalErr != nil {
			return internalErr
		}

		for _, sprintID := range sprintIDs {
			deleteSprintTaskRelationMutation := mutation.NewDeleteSprintTaskRelation(
				t.logger,
				t.stateSyncer,
				t.sprintTaskRelationDao,
				sprintID,
				task,
			)
			rtTx.AppendMutation(deleteSprintTaskRelationMutation)
			internalErr = deleteSprintTaskRelationMutation.Execute(ct, tx)
			if internalErr != nil {
				return internalErr
			}

			internalErr = t.storyTaskRelationDao.DeleteStoryTaskRelationsByTaskID(ct, tx, taskID)
			if internalErr != nil {
				return internalErr
			}
		}

		awaitForTaskIDs, internalErr := t.taskAwaitForRelationDao.FindAwaitForTaskIDsWithTx(ct, tx, taskID)
		if internalErr != nil {
			return internalErr
		}

		for _, awaitForTaskID := range awaitForTaskIDs {
			deleteTaskAwaitForRelationMutation := mutation.NewDeleteTaskAwaitForRelation(
				t.logger,
				t.stateSyncer,
				t.taskAwaitForRelationDao,
				task,
				awaitForTaskID,
			)

			rtTx.AppendMutation(deleteTaskAwaitForRelationMutation)
			internalErr = deleteTaskAwaitForRelationMutation.Execute(ct, tx)
			if internalErr != nil {
				return internalErr
			}
		}

		awaitingTaskIDs, internalErr := t.taskAwaitForRelationDao.FindAwaitingTaskIDsWithTx(ct, tx, taskID)
		if internalErr != nil {
			return internalErr
		}

		awaitingTasks, internalErr := t.taskDao.FindTasksByIDsWithTx(ct, tx, awaitingTaskIDs)
		if internalErr != nil {
			return internalErr
		}

		for _, awaitingTask := range awaitingTasks {
			deleteTaskAwaitForRelationMutation := mutation.NewDeleteTaskAwaitForRelation(
				t.logger,
				t.stateSyncer,
				t.taskAwaitForRelationDao,
				awaitingTask,
				taskID,
			)
			rtTx.AppendMutation(deleteTaskAwaitForRelationMutation)
			internalErr = deleteTaskAwaitForRelationMutation.Execute(ct, tx)
			if internalErr != nil {
				return internalErr
			}
		}

		internalErr = t.tryIncreaseBandwidth(ct, tx, rtTx, taskID, task.OwnerUserID, task.Effort)
		if internalErr != nil {
			return internalErr
		}

		deleteTaskMutation := mutation.NewDeleteTask(
			t.logger,
			t.stateSyncer,
			t.taskDao,
			task,
		)
		rtTx.AppendMutation(deleteTaskMutation)
		internalErr = deleteTaskMutation.Execute(ct, tx)
		if internalErr != nil {
			return internalErr
		}

		internalErr = t.threadDao.DeleteThread(ct, tx, task.CommentsThreadID)
		if internalErr != nil {
			return internalErr
		}

		return nil
	})

	// TODO: clean up resource relations in authorization service
	return task, err
}

func (t Task) moveTaskToUpcoming(ct context.Context, tx *cloudTransaction.Transaction, rtTx *realtime.Transaction, taskID uint64, autoPauseTask bool) (entity.Task, *errs.Error) {
	task, err := t.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
	if err != nil {
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

	now := time.Now().UTC()
	task.UpdatedAt = &now
	updateTaskMutation := mutation.NewUpdateTask(
		t.logger,
		t.stateSyncer,
		t.taskDao,
		task,
	)
	rtTx.AppendMutation(updateTaskMutation)
	internalErr := updateTaskMutation.Execute(ct, tx)
	if internalErr != nil {
		return entity.Task{}, internalErr
	}

	return task, err
}

func (t Task) MoveTaskToUpcoming(ct context.Context, taskID uint64, autoPauseTask bool) (entity.Task, *errs.Error) {
	if t.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.Task{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewMoveBetweenContainersInTaskQuery(userID, taskID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Task{}, err
		}

		if !hasPermission {
			return entity.Task{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var task entity.Task
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		task, err = t.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
		if err != nil {
			return err
		}

		if autoPauseTask {
			switch task.Status {
			case entity.TaskStatusInProgress, entity.TaskStatusPaused:
				task.Status = entity.TaskStatusPaused
			}
		} else {
			task.Status = entity.TaskStatusTodo
		}

		now := time.Now().UTC()
		task.UpdatedAt = &now
		updateTaskMutation := mutation.NewUpdateTask(
			t.logger,
			t.stateSyncer,
			t.taskDao,
			task,
		)
		rtTx.AppendMutation(updateTaskMutation)
		err = updateTaskMutation.Execute(ct, tx)
		if err != nil {
			return err
		}

		return nil
	})

	return task, err
}

func (t Task) MoveTaskToInProgress(ct context.Context, taskID uint64) (entity.Task, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return entity.Task{}, errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if t.featureToggles.EnableAuthorization {
		query := authorization.NewMoveBetweenContainersInTaskQuery(userID, taskID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Task{}, err
		}

		if !hasPermission {
			return entity.Task{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var task entity.Task
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		task, err = t.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
		if err != nil {
			return err
		}

		//tasks, err := t.taskDao.FindTasksByTeamIDWithTx(ct, tx, task.OwningTeamID)
		//if err != nil {
		//	return err
		//}

		if task.OwnerUserID == nil {
			task.OwnerUserID = &userID
		}

		now := time.Now().UTC()

		// TODO: enable based on team's setting
		//inProgressTasks := collect.Filter(tasks, func(eachTask entity.Task) bool {
		//	if eachTask.OwnerUserID == nil {
		//		return false
		//	}
		//
		//	if *eachTask.OwnerUserID != *task.OwnerUserID {
		//		return false
		//	}
		//
		//	return eachTask.Status == entity.TaskStatusInProgress
		//})
		//if len(inProgressTasks) > 0 {
		//	inProgressTask := inProgressTasks[0]
		//	inProgressTask.Status = entity.TaskStatusPaused
		//	inProgressTask.UpdatedAt = &now
		//	updateTaskMutation := mutation.NewUpdateTask(
		//		t.logger,
		//		t.stateSyncer,
		//		t.taskDao,
		//		inProgressTask,
		//	)
		//	rtTx.AppendMutation(updateTaskMutation)
		//	err = updateTaskMutation.Execute(ct, tx)
		//	if err != nil {
		//		return err
		//	}
		//}

		task.Status = entity.TaskStatusInProgress
		task.UpdatedAt = &now
		updateTaskMutation := mutation.NewUpdateTask(
			t.logger,
			t.stateSyncer,
			t.taskDao,
			task,
		)
		rtTx.AppendMutation(updateTaskMutation)
		err = updateTaskMutation.Execute(ct, tx)
		if err != nil {
			return err
		}

		return nil
	})

	return task, err
}

func (t Task) MoveTaskToDelivered(ct context.Context, taskID uint64) (entity.Task, *errs.Error) {
	if t.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.Task{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewMoveBetweenContainersInTaskQuery(userID, taskID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Task{}, err
		}

		if !hasPermission {
			return entity.Task{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var task entity.Task
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		task, err = t.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
		if err != nil {
			return err
		}

		task.Status = entity.TaskStatusDelivered
		now := time.Now().UTC()
		task.UpdatedAt = &now
		task.DeliveredAt = &now
		updateTaskMutation := mutation.NewUpdateTask(
			t.logger,
			t.stateSyncer,
			t.taskDao,
			task,
		)
		rtTx.AppendMutation(updateTaskMutation)
		err = updateTaskMutation.Execute(ct, tx)
		if err != nil {
			return err
		}

		awaitingTaskIDs, err := t.taskAwaitForRelationDao.FindAwaitingTaskIDsWithTx(ct, tx, taskID)
		if err != nil {
			return err
		}

		for _, awaitingTaskID := range awaitingTaskIDs {
			var awaitForTaskIDs []uint64
			awaitForTaskIDs, err = t.taskAwaitForRelationDao.FindAwaitForTaskIDsWithTx(ct, tx, awaitingTaskID)
			if err != nil {
				return err
			}

			var awaitForTasks []entity.Task
			awaitForTasks, err = t.taskDao.FindTasksByIDsWithTx(ct, tx, awaitForTaskIDs)
			if err != nil {
				return err
			}

			awaitForTasks = collect.Filter(awaitForTasks, func(awaitForTask entity.Task) bool {
				return awaitForTask.Status != entity.TaskStatusDelivered
			})
			if len(awaitForTasks) == 0 {
				_, err = t.moveTaskToUpcoming(ct, tx, rtTx, awaitingTaskID, false)
				if err != nil {
					return err
				}
			}
		}

		return nil
	})

	return task, err
}

func (t Task) MoveTaskToBlocked(ct context.Context, taskID uint64, reason string) (entity.Task, *errs.Error) {
	if t.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.Task{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewMoveBetweenContainersInTaskQuery(userID, taskID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Task{}, err
		}

		if !hasPermission {
			return entity.Task{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	panic("implement me")
}

func (t Task) AddAwaitForTask(ct context.Context, awaitingTaskID uint64, awaitForTaskId uint64) (entity.Task, *errs.Error) {
	if t.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.Task{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewAddAwaitForTaskInTaskQuery(userID, awaitingTaskID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Task{}, err
		}

		if !hasPermission {
			return entity.Task{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var task entity.Task
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		task, err = t.taskDao.FindTaskByIDWithTx(ct, tx, awaitingTaskID)
		if err != nil {
			return err
		}

		var awaitForTask entity.Task
		awaitForTask, err = t.taskDao.FindTaskByIDWithTx(ct, tx, awaitForTaskId)
		if err != nil {
			return err
		}

		if !awaitableTaskStatuses[task.Status] {
			return errs.NewError(errs.InvalidOperation, fmt.Sprintf("task must be awaitable: taskID=%v", awaitingTaskID))
		}

		now := time.Now().UTC()
		taskAwaitForRelation := entity.TaskAwaitForRelation{
			AwaitingTaskID: awaitingTaskID,
			AwaitForTaskID: awaitForTask.ID,
			CreatedAt:      now,
		}
		createTaskAwaitForRelationMutation := mutation.NewCreateTaskAwaitForRelation(
			t.logger,
			t.stateSyncer,
			t.taskAwaitForRelationDao,
			t.taskDao,
			taskAwaitForRelation,
		)
		rtTx.AppendMutation(createTaskAwaitForRelationMutation)
		err = createTaskAwaitForRelationMutation.Execute(ct, tx)
		if err != nil {
			return err
		}

		task.Status = entity.TaskStatusAwaiting
		task.UpdatedAt = &now
		updateTaskMutation := mutation.NewUpdateTask(
			t.logger,
			t.stateSyncer,
			t.taskDao,
			task,
		)
		rtTx.AppendMutation(updateTaskMutation)
		err = updateTaskMutation.Execute(ct, tx)
		if err != nil {
			return err
		}

		return nil
	})

	return task, err
}

func (t Task) RemoveAwaitForTask(ct context.Context, awaitingTaskID uint64, awaitForTaskId uint64) (entity.Task, *errs.Error) {
	if t.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.Task{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewRemoveAwaitForTaskInTaskQuery(userID, awaitingTaskID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Task{}, err
		}

		if !hasPermission {
			return entity.Task{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var task entity.Task
	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		task, err = t.taskDao.FindTaskByIDWithTx(ct, tx, awaitingTaskID)
		if err != nil {
			return err
		}

		var awaitForTask entity.Task
		awaitForTask, err = t.taskDao.FindTaskByIDWithTx(ct, tx, awaitForTaskId)
		if err != nil {
			return err
		}

		if task.Status != entity.TaskStatusAwaiting {
			return errs.NewError(errs.InvalidOperation, fmt.Sprintf("task must be awaitable: taskID=%v", awaitingTaskID))
		}

		deleteTaskAwaitForRelationMutation := mutation.NewDeleteTaskAwaitForRelation(
			t.logger,
			t.stateSyncer,
			t.taskAwaitForRelationDao,
			task,
			awaitForTask.ID,
		)
		rtTx.AppendMutation(deleteTaskAwaitForRelationMutation)
		err = deleteTaskAwaitForRelationMutation.Execute(ct, tx)
		if err != nil {
			return err
		}

		awaitForTaskIds, err := t.taskAwaitForRelationDao.FindAwaitForTaskIDsWithTx(ct, tx, awaitingTaskID)
		if err != nil {
			return err
		}

		if len(awaitForTaskIds) == 0 {
			task, err = t.moveTaskToUpcoming(ct, tx, rtTx, awaitingTaskID, false)
			if err != nil {
				return err
			}
		}

		return nil
	})

	return task, err
}

func (t Task) StartDraggingTask(ct context.Context, taskID uint64, clientID uint64) *errs.Error {
	userID, ok := ctx.UserIDFromContext(ct)
	if !ok {
		return errs.NewError(errs.Unauthenticated, "user ID not found")
	}

	if t.featureToggles.EnableAuthorization {
		query := authorization.NewMoveBetweenContainersInTaskQuery(userID, taskID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return err
		}

		if !hasPermission {
			return errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		task, err := t.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
		if err != nil {
			return err
		}

		taskActivity := entity.TaskActivity{
			TaskID: taskID,
			TeamID: task.OwningTeamID,
			DragTaskActivity: entity.DragTaskActivity{
				IsDragging: true,
				Client: &entity.Client{
					ID:     clientID,
					UserID: userID,
				},
			}}

		updateTaskActivityMutation := mutation.NewUpdateTaskActivity(
			t.logger,
			t.stateSyncer,
			t.activityCache,
			taskActivity,
		)
		rtTx.AppendMutation(updateTaskActivityMutation)
		err = updateTaskActivityMutation.Execute(ct, tx)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func (t Task) StopDraggingTask(ct context.Context, taskID uint64, clientID uint64) *errs.Error {
	if t.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewMoveBetweenContainersInTaskQuery(userID, taskID)
		hasPermission, err := t.authorizer.HasPermission(ct, query)
		if err != nil {
			return err
		}

		if !hasPermission {
			return errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	txCtx := transaction.NewTransactionsContext(
		t.logger,
		t.transactionFactory,
		t.stateSyncer,
		ct,
	)
	err := txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		task, err := t.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
		if err != nil {
			return err
		}

		taskActivity := entity.TaskActivity{
			TaskID: taskID,
			TeamID: task.OwningTeamID,
			DragTaskActivity: entity.DragTaskActivity{
				IsDragging: false,
				Client:     nil,
			}}

		updateTaskActivityMutation := mutation.NewUpdateTaskActivity(
			t.logger,
			t.stateSyncer,
			t.activityCache,
			taskActivity,
		)
		rtTx.AppendMutation(updateTaskActivityMutation)
		err = updateTaskActivityMutation.Execute(ct, tx)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}

func NewTask(
	logger telemetry.Logger,
	cloudClientRegistry *client.Registry,
	authorizer client.Authorizer,
	featureToggles feature.Toggles,
	stateSyncer *realtime.StateSyncer,
	transactionFactory cloudTransaction.Factory,
	activityCache cache.Activity,
	taskDao dao.Task,
	threadDao dao.Thread,
	sprintDao dao.Sprint,
	taskAwaitForRelationDao dao.TaskAwaitForRelation,
	sprintParticipantDao dao.SprintParticipant,
	sprintTaskRelationDao dao.SprintTaskRelation,
	storyTaskRelationDao dao.StoryTaskRelation,
) Task {
	return Task{
		logger:                  logger,
		cloudClientRegistry:     cloudClientRegistry,
		authorizer:              authorizer,
		featureToggles:          featureToggles,
		stateSyncer:             stateSyncer,
		transactionFactory:      transactionFactory,
		activityCache:           activityCache,
		taskDao:                 taskDao,
		threadDao:               threadDao,
		sprintDao:               sprintDao,
		taskAwaitForRelationDao: taskAwaitForRelationDao,
		sprintParticipantDao:    sprintParticipantDao,
		sprintTaskRelationDao:   sprintTaskRelationDao,
		storyTaskRelationDao:    storyTaskRelationDao,
	}
}
