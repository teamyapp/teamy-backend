package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

const timePerWeek = 7 * 24 * time.Hour

const (
	TooManySprints errs.ErrorCode = "TooManySprints"
)

type CreateSprintInput struct {
	StartAt time.Time
	EndAt   time.Time
}

type Sprint struct {
	dataCollector           telemetry.DataCollector
	cloudClientRegistry     *cloudAPI.ClientRegistry
	stateSyncer             *realtime.StateSyncer
	authorizer              Authorizer
	taskDao                 dao.Task
	taskDaoV2               daov2.Task
	sprintDao               dao.Sprint
	sprintDaoV2             daov2.Sprint
	sprintTaskRelationDao   dao.SprintTaskRelation
	sprintTaskRelationDaoV2 daov2.SprintTaskRelation
	sprintParticipantDao    dao.SprintParticipant
	sprintParticipantDaoV2  daov2.SprintParticipant
	teamMemberDao           dao.TeamMember
	teamMemberDaoV2         daov2.TeamMember
	threadDaoV2             daov2.Thread
	db                      *sql.DB
}

func (s Sprint) FindSprintsInTeam(ct context.Context, teamID uint64, filter *SprintFilter) ([]entity.Sprint, *errs.Error) {
	var sprints []entity.Sprint
	txCtx := TransactionsContext{
		dataCollector: s.dataCollector,
		db:            s.db,
		stateSyncer:   s.stateSyncer,
		ct:            ct,
	}
	err := txCtx.withTransactions(true, func(sqlTx *sql.Tx, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		sprints, err = s.sprintDaoV2.FindSprintsByTeamID(ct, sqlTx, teamID)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		if filter != nil {
			sprints = filterSprints(sprints, *filter)
		}

		return nil
	})

	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, err
	}

	return sprints, nil
}

func (s Sprint) FindParticipantsInSprint(ct context.Context, sprintID uint64) ([]entity.SprintParticipant, *errs.Error) {
	var participants []entity.SprintParticipant
	txCtx := TransactionsContext{
		dataCollector: s.dataCollector,
		db:            s.db,
		stateSyncer:   s.stateSyncer,
		ct:            ct,
	}
	err := txCtx.withTransactions(true, func(sqlTx *sql.Tx, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		participants, err = s.sprintParticipantDaoV2.FindParticipantsBySprintID(ct, sqlTx, sprintID)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		return nil
	})

	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, err
	}

	return participants, nil
}

func (s Sprint) FindSprints(ct context.Context, filter *SprintFilter) ([]entity.Sprint, *errs.Error) {
	var sprints []entity.Sprint
	txCtx := TransactionsContext{
		dataCollector: s.dataCollector,
		db:            s.db,
		stateSyncer:   s.stateSyncer,
		ct:            ct,
	}
	err := txCtx.withTransactions(true, func(sqlTx *sql.Tx, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		sprints, err = s.sprintDaoV2.FindAllSprints(ct, sqlTx)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		if filter != nil {
			sprints = filterSprints(sprints, *filter)
		}

		return nil
	})

	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, err
	}

	return sprints, nil
}

func (s Sprint) FindCurrentSprint(ct context.Context, teamID uint64) (entity.Sprint, *errs.Error) {
	var sprint entity.Sprint
	txCtx := TransactionsContext{
		dataCollector: s.dataCollector,
		db:            s.db,
		stateSyncer:   s.stateSyncer,
		ct:            ct,
	}
	err := txCtx.withTransactions(true, func(sqlTx *sql.Tx, rtTx *realtime.Transaction) *errs.Error {
		sprints, err := s.sprintDaoV2.FindSprintsByTeamID(ct, sqlTx, teamID)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		now := time.Now().UTC()
		sprints = collect.Filter(sprints, func(sprint entity.Sprint) bool {
			if now.Before(sprint.StartAt.UTC()) || now.After(sprint.EndAt.UTC()) {
				return false
			}

			return true
		})
		if len(sprints) < 1 {
			internalErr := &errs.Error{
				Code:    errs.Unauthenticated,
				Message: fmt.Sprintf("team has no active sprint: teamID=%v, currentTime=%v", teamID, now.UTC()),
			}
			s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return internalErr
		}

		if len(sprints) > 1 {
			internalErr := &errs.Error{
				Code:    TooManySprints,
				Message: fmt.Sprintf("team has more than one current sprint: teamID=%v, currentTime=%v", teamID, now.UTC()),
			}
			s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return internalErr
		}

		sprint = sprints[0]
		return nil
	})

	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Sprint{}, err
	}

	return sprint, nil
}

func (s Sprint) FindCurrentAndFutureSprints(ct context.Context, teamID uint64) ([]entity.Sprint, *errs.Error) {
	var sprints []entity.Sprint
	txCtx := TransactionsContext{
		dataCollector: s.dataCollector,
		db:            s.db,
		stateSyncer:   s.stateSyncer,
		ct:            ct,
	}
	err := txCtx.withTransactions(true, func(sqlTx *sql.Tx, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		sprints, err = s.sprintDaoV2.FindSprintsByTeamID(ct, sqlTx, teamID)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		now := time.Now().UTC()
		sprints = collect.Filter(sprints, func(sprint entity.Sprint) bool {
			if sprint.EndAt.UTC().Before(now) {
				return false
			}

			return true
		})
		return nil
	})

	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, err
	}

	return sprints, nil
}

func (s Sprint) FindSprintByID(ct context.Context, sprintID uint64) (entity.Sprint, *errs.Error) {
	var sprint entity.Sprint
	txCtx := TransactionsContext{
		dataCollector: s.dataCollector,
		db:            s.db,
		stateSyncer:   s.stateSyncer,
		ct:            ct,
	}
	err := txCtx.withTransactions(true, func(sqlTx *sql.Tx, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		sprint, err = s.sprintDaoV2.FindSprintByID(ct, sqlTx, sprintID)
		return err
	})

	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Sprint{}, err
	}

	return sprint, nil
}

func (s Sprint) CreateSprint(ct context.Context, teamID uint64, input CreateSprintInput) (entity.Sprint, *errs.Error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			internalErr := &errs.Error{
				Code:    errs.Unauthenticated,
				Message: "user ID not found",
			}
			s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.Sprint{}, internalErr
		}

		query := authorization.NewTeamCreateSprintQuery(userID, teamID)
		hasPermission, err := s.authorizer.hasPermission(ct, query)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.Sprint{}, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return entity.Sprint{}, internalErr
		}
	}

	genSprintIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "sprintID"}
	genSprintIDRes, rpcErr := s.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genSprintIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Sprint{}, internalErr
	}

	var sprint entity.Sprint
	txCtx := TransactionsContext{
		dataCollector: s.dataCollector,
		db:            s.db,
		stateSyncer:   s.stateSyncer,
		ct:            ct,
	}
	err := txCtx.withTransactions(false, func(sqlTx *sql.Tx, rtTx *realtime.Transaction) *errs.Error {
		sprint = entity.Sprint{
			ID:           genSprintIDRes.UniqueNumber,
			StartAt:      input.StartAt.UTC(),
			EndAt:        input.EndAt.UTC(),
			CreatedAt:    time.Now().UTC(),
			OwningTeamID: teamID,
		}
		err := s.sprintDaoV2.CreateSprint(ct, sqlTx, sprint)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		teamMembers, err := s.teamMemberDaoV2.FindTeamMembersByTeamID(ct, sqlTx, teamID)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		sprintLength := input.EndAt.UTC().Sub(input.StartAt.UTC())
		numOfWeeks := sprintLength / timePerWeek
		// TODO: fetch from team settings

		for _, teamMember := range teamMembers {
			totalBandwidth := teamMember.WeeklyBandwidth * numOfWeeks
			participant := entity.SprintParticipant{
				SprintID:        sprint.ID,
				UserID:          teamMember.UserID,
				TotalBandwidth:  totalBandwidth,
				UnusedBandwidth: totalBandwidth,
				CreatedAt:       time.Now(),
			}
			createSprintParticipantMutation := mutation.NewCreateSprintParticipantMutation(
				s.dataCollector,
				s.stateSyncer,
				s.sprintParticipantDao,
				s.sprintParticipantDaoV2,
				s.sprintDao,
				s.sprintDaoV2,
				participant)
			rtTx.AppendMutation(createSprintParticipantMutation)
			err = createSprintParticipantMutation.ExecuteV2(ct, sqlTx)
			if err != nil {
				s.dataCollector.Logger.ErrorWithContext(ct, err)
				return err
			}
		}

		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		return nil
	})

	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Sprint{}, err
	}

	// TODO(yuhang): if failed to register/assign resource, there will be inconsistent state. Cross-service transaction
	// protection will be covered in stage 2
	if feature.EnableAuthorization {
		err = s.authorizer.registerResource(ct, authorization.SprintResourceType, sprint.ID)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.Sprint{}, err
		}

		err = s.authorizer.assignParentResource(ct, authorization.SprintResourceType, sprint.ID, authorization.TeamResourceType, sprint.OwningTeamID)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.Sprint{}, err
		}
	}

	return sprint, nil
}

func (s Sprint) DeleteSprint(ct context.Context, sprintID uint64) (entity.Sprint, *errs.Error) {
	var sprint entity.Sprint
	txCtx := TransactionsContext{
		dataCollector: s.dataCollector,
		db:            s.db,
		stateSyncer:   s.stateSyncer,
		ct:            ct,
	}
	err := txCtx.withTransactions(false, func(sqlTx *sql.Tx, rtTx *realtime.Transaction) *errs.Error {
		taskIds, err := s.sprintTaskRelationDaoV2.FindTaskIDsBySprintID(ct, sqlTx, sprintID)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		for _, taskId := range taskIds {
			_, err = s.removeTaskFromSprint(ct, sqlTx, rtTx, sprintID, taskId)
			if err != nil {
				s.dataCollector.Logger.ErrorWithContext(ct, err)
				return err
			}
		}

		participantUserIDs, err := s.sprintParticipantDaoV2.FindParticipantIDsBySprintID(ct, sqlTx, sprintID)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		sprint, err = s.sprintDaoV2.FindSprintByID(ct, sqlTx, sprintID)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		for _, participantUserID := range participantUserIDs {
			deleteSprintParticipantMutation := mutation.NewDeleteSprintParticipantMutation(
				s.dataCollector,
				s.stateSyncer,
				s.sprintParticipantDao,
				s.sprintParticipantDaoV2,
				s.sprintDao,
				s.sprintDaoV2,
				participantUserID,
				sprintID)
			rtTx.AppendMutation(deleteSprintParticipantMutation)
			err = deleteSprintParticipantMutation.ExecuteV2(ct, sqlTx)
			if err != nil {
				s.dataCollector.Logger.ErrorWithContext(ct, err)
				return err
			}

			// we need to prepare notifier in advance since sprint will be actually deleted later
			deleteSprintParticipantMutation.PrepareClientNotifiers(ct, sqlTx)
		}

		return s.sprintDaoV2.DeleteSprint(ct, sqlTx, sprintID)
	})

	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Sprint{}, err
	}

	return sprint, nil
}

func (s Sprint) AddTaskToSprint(ct context.Context, sprintID uint64, taskID uint64) (entity.Task, *errs.Error) {
	var task entity.Task
	txCtx := TransactionsContext{
		dataCollector: s.dataCollector,
		db:            s.db,
		stateSyncer:   s.stateSyncer,
		ct:            ct,
	}
	err := txCtx.withTransactions(false, func(sqlTx *sql.Tx, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		task, err = s.taskDaoV2.FindTaskByID(ct, sqlTx, taskID)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		sprint, err := s.sprintDaoV2.FindSprintByID(ct, sqlTx, sprintID)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		if sprint.OwningTeamID != task.OwningTeamID {
			internalErr := &errs.Error{
				Code:    errs.InvalidArgument,
				Message: fmt.Sprintf("sprint and task must belong to the same team: sprintID=%v, taskID=%v", sprintID, taskID),
			}
			s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return internalErr
		}

		relation := entity.SprintTaskRelation{
			SprintID:  sprintID,
			TaskID:    taskID,
			CreatedAt: time.Now().UTC(),
		}
		createSprintTaskRelationMutation := mutation.NewCreateSprintTaskRelationMutation(
			s.dataCollector,
			s.stateSyncer,
			s.sprintTaskRelationDao,
			s.sprintTaskRelationDaoV2,
			s.sprintDao,
			s.sprintDaoV2,
			relation)
		rtTx.AppendMutation(createSprintTaskRelationMutation)
		err = createSprintTaskRelationMutation.ExecuteV2(ct, sqlTx)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		if !task.IsPlanned {
			task.IsPlanned = true
			updateTaskMutation := mutation.NewUpdateTaskMutation(
				s.dataCollector,
				s.stateSyncer,
				s.taskDao,
				s.taskDaoV2,
				task)
			rtTx.AppendMutation(updateTaskMutation)
			err = updateTaskMutation.ExecuteV2(ct, sqlTx)
			if err != nil {
				s.dataCollector.Logger.ErrorWithContext(ct, err)
				return err
			}
		}

		err = s.tryReduceBandwidth(ct, sqlTx, rtTx, sprintID, task)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		return nil
	})

	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	return task, nil
}

func (s Sprint) CopyTasksToSprint(ct context.Context, toSprintID uint64, taskIDs []uint64) ([]entity.Task, *errs.Error) {
	var sprint entity.Sprint
	var err *errs.Error
	txCtx := TransactionsContext{
		dataCollector: s.dataCollector,
		db:            s.db,
		stateSyncer:   s.stateSyncer,
		ct:            ct,
	}
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			internalErr := &errs.Error{
				Code:    errs.Unauthenticated,
				Message: "user ID not found",
			}
			s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return nil, internalErr
		}

		// TODO(yuhang): better call non-transactional query here, will update after we add those functions in daoV2
		txCtx.withTransactions(true, func(sqlTx *sql.Tx, rtTx *realtime.Transaction) *errs.Error {
			sprint, err = s.sprintDaoV2.FindSprintByID(ct, sqlTx, toSprintID)
			return err
		})
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return nil, err
		}

		query := authorization.NewTeamCloneTaskQuery(userID, sprint.OwningTeamID)
		hasPermission, err := s.authorizer.hasPermission(ct, query)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return nil, err
		}

		if !hasPermission {
			internalErr := &errs.Error{
				Code:    errs.PermissionDenied,
				Message: fmt.Sprintf("permission denied: authorization query=%v", query),
			}
			s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return []entity.Task{}, internalErr
		}
	}

	var tasks []entity.Task
	var newTaskIDs []uint64
	var newThreadIDs []uint64
	// TODO(yuhang): these genID requests should be batched in one
	for range taskIDs {
		genTaskIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "taskID"}
		genTaskIDRes, rpcErr := s.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTaskIDReq)
		if rpcErr != nil {
			internalErr := errs.FromGRPCErr(rpcErr)
			s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return nil, internalErr
		}

		genThreadIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "threadID"}
		genThreadIDRes, rpcErr := s.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genThreadIDReq)
		if rpcErr != nil {
			internalErr := errs.FromGRPCErr(rpcErr)
			s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
			return nil, internalErr
		}

		newTaskIDs = append(newTaskIDs, genTaskIDRes.UniqueNumber)
		newThreadIDs = append(newThreadIDs, genThreadIDRes.UniqueNumber)
	}

	err = txCtx.withTransactions(false, func(sqlTx *sql.Tx, rtTx *realtime.Transaction) *errs.Error {
		for idx, taskID := range taskIDs {
			var task entity.Task
			task, err = s.copyTaskToSprint(ct, sqlTx, rtTx, toSprintID, taskID, newTaskIDs[idx], newThreadIDs[idx])
			if err != nil {
				s.dataCollector.Logger.ErrorWithContext(ct, err)
				continue
			}

			tasks = append(tasks, task)
		}

		return nil
	})

	return tasks, nil
}

func (s Sprint) MoveTasksToSprint(ct context.Context, fromSprintID uint64, toSprintID uint64, taskIDs []uint64) ([]entity.Task, *errs.Error) {
	var tasks []entity.Task
	txCtx := TransactionsContext{
		dataCollector: s.dataCollector,
		db:            s.db,
		stateSyncer:   s.stateSyncer,
		ct:            ct,
	}
	err := txCtx.withTransactions(false, func(sqlTx *sql.Tx, rtTx *realtime.Transaction) *errs.Error {
		for _, taskID := range taskIDs {
			task, err := s.moveTaskToSprint(ct, sqlTx, rtTx, fromSprintID, toSprintID, taskID)

			if err != nil {
				s.dataCollector.Logger.ErrorWithContext(ct, err)
				continue
			}

			tasks = append(tasks, task)
		}

		return nil
	})

	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return nil, err
	}

	return tasks, nil
}

func (s Sprint) RemoveTaskFromSprint(ct context.Context, sprintID uint64, taskID uint64) (entity.Task, *errs.Error) {
	var task entity.Task
	txCtx := TransactionsContext{
		dataCollector: s.dataCollector,
		db:            s.db,
		stateSyncer:   s.stateSyncer,
		ct:            ct,
	}
	err := txCtx.withTransactions(false, func(sqlTx *sql.Tx, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		task, err = s.removeTaskFromSprint(ct, sqlTx, rtTx, sprintID, taskID)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		return nil
	})

	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	return task, nil
}

func (s Sprint) copyTaskToSprint(ct context.Context, sqlTx *sql.Tx, rtTx *realtime.Transaction, toSprintID uint64, taskID uint64, newTaskID uint64, newThreadID uint64) (entity.Task, *errs.Error) {
	task, err := s.taskDaoV2.FindTaskByID(ct, sqlTx, taskID)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	sprint, err := s.sprintDaoV2.FindSprintByID(ct, sqlTx, toSprintID)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	if sprint.OwningTeamID != task.OwningTeamID {
		internalErr := &errs.Error{
			Code:    errs.InvalidArgument,
			Message: fmt.Sprintf("sprint and task must belong to the same team: sprintID=%v, taskID=%v", toSprintID, newTaskID),
		}
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Task{}, internalErr
	}

	err = s.threadDaoV2.CreateThread(ct, sqlTx, newThreadID)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	newTask := entity.Task{
		ID:               newTaskID,
		Goal:             task.Goal,
		Context:          task.Context,
		Status:           task.Status,
		IsPlanned:        task.IsPlanned,
		CreatorUserID:    task.CreatorUserID,
		OwningTeamID:     task.OwningTeamID,
		Effort:           task.Effort,
		OwnerUserID:      task.OwnerUserID,
		CommentsThreadID: newThreadID,
		CreatedAt:        time.Now().UTC(),
		DueAt:            task.DueAt,
		DeliveredAt:      task.DeliveredAt,
	}

	createTaskMutation := mutation.NewCreateTaskMutation(
		s.dataCollector,
		s.stateSyncer,
		s.taskDao,
		s.taskDaoV2,
		newTask,
	)
	err = createTaskMutation.ExecuteV2(ct, sqlTx)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	rtTx.AppendMutation(createTaskMutation)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	relation := entity.SprintTaskRelation{
		SprintID:  toSprintID,
		TaskID:    newTaskID,
		CreatedAt: time.Now().UTC(),
	}
	createSprintTaskRelationMutation := mutation.NewCreateSprintTaskRelationMutation(
		s.dataCollector,
		s.stateSyncer,
		s.sprintTaskRelationDao,
		s.sprintTaskRelationDaoV2,
		s.sprintDao,
		s.sprintDaoV2,
		relation)
	rtTx.AppendMutation(createSprintTaskRelationMutation)
	err = createSprintTaskRelationMutation.ExecuteV2(ct, sqlTx)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	if !task.IsPlanned {
		task.IsPlanned = true
		updateTaskMutation := mutation.NewUpdateTaskMutation(
			s.dataCollector,
			s.stateSyncer,
			s.taskDao,
			s.taskDaoV2,
			task)
		rtTx.AppendMutation(updateTaskMutation)
		err = updateTaskMutation.ExecuteV2(ct, sqlTx)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.Task{}, err
		}
	}

	err = s.tryReduceBandwidth(ct, sqlTx, rtTx, toSprintID, task)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	return newTask, nil
}

func (s Sprint) moveTaskToSprint(ct context.Context, sqlTx *sql.Tx, rtTx *realtime.Transaction, fromSprintID uint64, toSprintID uint64, taskID uint64) (entity.Task, *errs.Error) {
	sprintIDs, err := s.sprintTaskRelationDaoV2.FindSprintIDsByTaskID(ct, sqlTx, taskID)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	foundSprintIDs := collect.Filter(sprintIDs, func(currSprintID uint64) bool {
		return currSprintID == fromSprintID
	})
	if len(foundSprintIDs) < 1 {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("relation not found: sprintID=%v, taskID=%v", fromSprintID, taskID),
		}
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Task{}, internalErr
	}

	err = s.sprintTaskRelationDaoV2.DeleteSprintTaskRelation(ct, sqlTx, fromSprintID, taskID)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	relation := entity.SprintTaskRelation{
		SprintID:  toSprintID,
		TaskID:    taskID,
		CreatedAt: time.Now().UTC(),
	}

	err = s.sprintTaskRelationDaoV2.CreateSprintTaskRelation(ct, sqlTx, relation)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	task, err := s.taskDaoV2.FindTaskByID(ct, sqlTx, taskID)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	err = s.tryIncreaseBandwidth(ct, sqlTx, rtTx, fromSprintID, task)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	err = s.tryReduceBandwidth(ct, sqlTx, rtTx, toSprintID, task)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	return task, nil
}

func (s Sprint) removeTaskFromSprint(ct context.Context, sqlTx *sql.Tx, rtTx *realtime.Transaction, sprintID uint64, taskID uint64) (entity.Task, *errs.Error) {
	sprintIDs, err := s.sprintTaskRelationDaoV2.FindSprintIDsByTaskID(ct, sqlTx, taskID)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	foundSprintIDs := collect.Filter(sprintIDs, func(currSprintID uint64) bool {
		return currSprintID == sprintID
	})
	if len(foundSprintIDs) < 1 {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("relation not found: sprintID=%v, taskID=%v", sprintID, taskID),
		}
		s.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Task{}, internalErr
	}

	task, err := s.taskDaoV2.FindTaskByID(ct, sqlTx, taskID)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	deleteSprintTaskRelationMutation := mutation.NewDeleteSprintTaskRelationMutation(
		s.dataCollector,
		s.stateSyncer,
		s.sprintTaskRelationDao,
		s.sprintTaskRelationDaoV2,
		sprintID,
		task,
	)
	rtTx.AppendMutation(deleteSprintTaskRelationMutation)
	err = deleteSprintTaskRelationMutation.ExecuteV2(ct, sqlTx)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	//if there is no other sprint that the task can move to,  put it into backlog
	if len(sprintIDs) <= 1 {
		task.IsPlanned = false
		updateTaskMutation := mutation.NewUpdateTaskMutation(
			s.dataCollector,
			s.stateSyncer,
			s.taskDao,
			s.taskDaoV2,
			task)
		rtTx.AppendMutation(updateTaskMutation)
		err = updateTaskMutation.ExecuteV2(ct, sqlTx)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return entity.Task{}, err
		}
	}

	err = s.tryIncreaseBandwidth(ct, sqlTx, rtTx, sprintID, task)
	if err != nil {
		s.dataCollector.Logger.ErrorWithContext(ct, err)
		return entity.Task{}, err
	}

	return task, nil
}

func (s Sprint) tryReduceBandwidth(ct context.Context, sqlTx *sql.Tx, rtTx *realtime.Transaction, sprintID uint64, task entity.Task) *errs.Error {
	if task.OwnerUserID != nil && task.Effort != nil {
		newSprintParticipant, err := s.sprintParticipantDaoV2.FindParticipant(ct, sqlTx, sprintID, *task.OwnerUserID)
		if err != nil {
			// TODO: this should be removed once the sprint participants are backfilled
			if err.Code == errs.NotFound {
				return nil
			}

			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		newSprintParticipant.UnusedBandwidth -= *task.Effort
		updateSprintParticipantMutation := mutation.NewUpdateSprintParticipantMutation(
			s.dataCollector,
			s.stateSyncer,
			s.sprintParticipantDao,
			s.sprintParticipantDaoV2,
			s.sprintDao,
			s.sprintDaoV2,
			newSprintParticipant)
		rtTx.AppendMutation(updateSprintParticipantMutation)
		err = updateSprintParticipantMutation.ExecuteV2(ct, sqlTx)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}
	}

	return nil
}

func (s Sprint) tryIncreaseBandwidth(ct context.Context, sqlTx *sql.Tx, rtTx *realtime.Transaction, sprintID uint64, task entity.Task) *errs.Error {
	if task.OwnerUserID != nil && task.Effort != nil {
		oldSprintParticipant, err := s.sprintParticipantDaoV2.FindParticipant(ct, sqlTx, sprintID, *task.OwnerUserID)
		if err != nil {
			// TODO: this should be removed once the sprint participants are backfilled
			if err.Code == errs.NotFound {
				return nil
			}

			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}

		oldSprintParticipant.UnusedBandwidth += *task.Effort
		updateSprintParticipantMutation := mutation.NewUpdateSprintParticipantMutation(
			s.dataCollector,
			s.stateSyncer,
			s.sprintParticipantDao,
			s.sprintParticipantDaoV2,
			s.sprintDao,
			s.sprintDaoV2,
			oldSprintParticipant)
		rtTx.AppendMutation(updateSprintParticipantMutation)
		err = updateSprintParticipantMutation.ExecuteV2(ct, sqlTx)
		if err != nil {
			s.dataCollector.Logger.ErrorWithContext(ct, err)
			return err
		}
	}

	return nil
}

func NewSprint(
	dataCollector telemetry.DataCollector,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	stateSyncer *realtime.StateSyncer,
	authorizer Authorizer,
	taskDao dao.Task,
	taskDaoV2 daov2.Task,
	sprintDao dao.Sprint,
	sprintDaoV2 daov2.Sprint,
	sprintTaskRelationDao dao.SprintTaskRelation,
	sprintTaskRelationDaoV2 daov2.SprintTaskRelation,
	sprintParticipantDao dao.SprintParticipant,
	sprintParticipantDaoV2 daov2.SprintParticipant,
	teamMemberDao dao.TeamMember,
	teamMemberDaoV2 daov2.TeamMember,
	threadDaoV2 daov2.Thread,
	db *sql.DB,
) Sprint {
	return Sprint{
		dataCollector:           dataCollector,
		cloudClientRegistry:     cloudClientRegistry,
		stateSyncer:             stateSyncer,
		authorizer:              authorizer,
		taskDao:                 taskDao,
		taskDaoV2:               taskDaoV2,
		sprintDao:               sprintDao,
		sprintDaoV2:             sprintDaoV2,
		sprintTaskRelationDao:   sprintTaskRelationDao,
		sprintTaskRelationDaoV2: sprintTaskRelationDaoV2,
		sprintParticipantDao:    sprintParticipantDao,
		sprintParticipantDaoV2:  sprintParticipantDaoV2,
		teamMemberDao:           teamMemberDao,
		teamMemberDaoV2:         teamMemberDaoV2,
		threadDaoV2:             threadDaoV2,
		db:                      db,
	}
}
