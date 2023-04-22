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
	"github.com/teamyapp/cloud/libs/transaction"
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
	logger                  telemetry.Logger
	cloudClientRegistry     *cloudAPI.ClientRegistry
	stateSyncer             *realtime.StateSyncer
	authorizer              Authorizer
	featureToggles          feature.Toggles
	transactionFactory      transaction.Factory
	taskDao                 dao.Task
	taskDaoV2               daov2.Task
	sprintDao               dao.Sprint
	sprintDaoV2             daov2.Sprint
	teamDao                 dao.Team
	teamDaoV2               daov2.Team
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
	sprints, err := s.sprintDaoV2.FindSprintsByTeamID(ct, teamID)
	if err != nil {
		return nil, err
	}

	if filter != nil {
		sprints = filterSprints(sprints, *filter)
	}

	return sprints, nil
}

func (s Sprint) FindParticipantsInSprint(ct context.Context, sprintID uint64) ([]entity.SprintParticipant, *errs.Error) {
	return s.sprintParticipantDaoV2.FindParticipantsBySprintID(ct, sprintID)
}

func (s Sprint) FindSprints(ct context.Context, filter *SprintFilter) ([]entity.Sprint, *errs.Error) {
	sprints, err := s.sprintDaoV2.FindAllSprints(ct)
	if err != nil {
		return nil, err
	}

	if filter != nil {
		sprints = filterSprints(sprints, *filter)
	}

	return sprints, nil
}

func (s Sprint) GetActiveSprint(ct context.Context, teamID uint64) (*entity.Sprint, *errs.Error) {
	if s.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewReadInTeamQuery(userID, teamID)
		hasPermission, err := s.authorizer.hasPermission(ct, query)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			return nil, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	txCtx := TransactionsContext{
		logger:             s.logger,
		transactionFactory: s.transactionFactory,
		stateSyncer:        s.stateSyncer,
		ct:                 ct,
	}
	var sprint *entity.Sprint
	err := txCtx.withTransactions(true, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		team, err := s.teamDaoV2.FindTeamByIDWithTx(ct, tx, teamID)
		if err != nil {
			return err
		}

		if team.ActiveSprintID == nil {
			return nil
		}

		sprintRes, err := s.sprintDaoV2.FindSprintByID(ct, *team.ActiveSprintID)
		if err != nil {
			return err
		}

		sprint = &sprintRes
		return nil
	})

	if err != nil {
		return nil, err
	}

	return sprint, nil
}

func (s Sprint) SetTeamActiveSprint(ct context.Context, teamID uint64, sprintID uint64) (entity.Team, *errs.Error) {
	if s.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.Team{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewUpdateInTeamQuery(userID, teamID)
		hasPermission, err := s.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.Team{}, err
		}

		if !hasPermission {
			return entity.Team{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	txCtx := TransactionsContext{
		logger:             s.logger,
		transactionFactory: s.transactionFactory,
		stateSyncer:        s.stateSyncer,
		ct:                 ct,
	}
	var team entity.Team
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		team, err = s.teamDaoV2.FindTeamByIDWithTx(ct, tx, teamID)
		if err != nil {
			return err
		}

		sprint, err := s.sprintDaoV2.FindSprintByIDWithTx(ct, tx, sprintID)

		if err != nil {
			return err
		}

		if sprint.OwningTeamID != teamID {
			return errs.NewError(errs.InvalidArgument, "Sprint does not belong to team")
		}

		team.ActiveSprintID = &sprintID
		updatedAt := time.Now().UTC()
		team.UpdatedAt = &updatedAt

		updateTeamMutation := mutation.NewUpdateTeam(
			s.logger,
			s.stateSyncer,
			s.teamDao,
			s.teamDaoV2,
			team,
		)
		rtTx.AppendMutation(updateTeamMutation)
		return updateTeamMutation.ExecuteV2(ct, tx)
	})

	if err != nil {
		return entity.Team{}, err
	}

	return team, nil
}

func (s Sprint) FindCurrentAndFutureSprints(ct context.Context, teamID uint64) ([]entity.Sprint, *errs.Error) {
	var sprints []entity.Sprint
	txCtx := TransactionsContext{
		logger:             s.logger,
		transactionFactory: s.transactionFactory,
		stateSyncer:        s.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(true, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		sprints, err = s.sprintDaoV2.FindSprintsByTeamIDWithTx(ct, tx, teamID)
		if err != nil {
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
		return nil, err
	}

	return sprints, nil
}

func (s Sprint) FindSprintByID(ct context.Context, sprintID uint64) (entity.Sprint, *errs.Error) {
	var sprint entity.Sprint
	txCtx := TransactionsContext{
		logger:             s.logger,
		transactionFactory: s.transactionFactory,
		stateSyncer:        s.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(true, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		sprint, err = s.sprintDaoV2.FindSprintByIDWithTx(ct, tx, sprintID)
		return err
	})

	if err != nil {
		return entity.Sprint{}, err
	}

	return sprint, nil
}

func (s Sprint) CreateSprint(ct context.Context, teamID uint64, input CreateSprintInput) (entity.Sprint, *errs.Error) {
	if s.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.Sprint{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewCreateSprintInTeamQuery(userID, teamID)
		hasPermission, err := s.authorizer.hasPermission(ct, query)
		if err != nil {
			return entity.Sprint{}, err
		}

		if !hasPermission {
			return entity.Sprint{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	genSprintIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "sprintID"}
	genSprintIDRes, rpcErr := s.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genSprintIDReq)
	if rpcErr != nil {
		internalErr := errs.FromGRPCErr(rpcErr)
		return entity.Sprint{}, internalErr
	}

	var sprint entity.Sprint
	txCtx := TransactionsContext{
		logger:             s.logger,
		transactionFactory: s.transactionFactory,
		stateSyncer:        s.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		sprint = entity.Sprint{
			ID:           genSprintIDRes.UniqueNumber,
			StartAt:      input.StartAt.UTC(),
			EndAt:        input.EndAt.UTC(),
			CreatedAt:    time.Now().UTC(),
			OwningTeamID: teamID,
		}

		createSprintMutation := mutation.NewCreateSprintMutation(
			s.logger,
			s.stateSyncer,
			s.sprintDaoV2,
			sprint)

		rtTx.AppendMutation(createSprintMutation)
		err := createSprintMutation.ExecuteV2(ct, tx)
		if err != nil {
			return err
		}

		teamMembers, err := s.teamMemberDaoV2.FindTeamMembersByTeamIDWithTx(ct, tx, teamID)
		if err != nil {
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
			createSprintParticipantMutation := mutation.NewCreateSprintParticipant(
				s.logger,
				s.stateSyncer,
				s.sprintParticipantDao,
				s.sprintParticipantDaoV2,
				s.sprintDao,
				s.sprintDaoV2,
				participant)
			rtTx.AppendMutation(createSprintParticipantMutation)
			err = createSprintParticipantMutation.ExecuteV2(ct, tx)
			if err != nil {
				return err
			}
		}

		return nil
	})

	if err != nil {
		return entity.Sprint{}, err
	}

	// TODO(yuhang): if failed to register/assign resource, there will be inconsistent state. Cross-service transaction
	// protection will be covered in stage 2
	if s.featureToggles.EnableAuthorization {
		err = s.authorizer.registerResource(ct, authorization.SprintResourceType, sprint.ID)
		if err != nil {
			return entity.Sprint{}, err
		}

		err = s.authorizer.assignParentResource(ct, authorization.SprintResourceType, sprint.ID, authorization.TeamResourceType, sprint.OwningTeamID)
		if err != nil {
			return entity.Sprint{}, err
		}
	}

	return sprint, nil
}

func (s Sprint) DeleteSprint(ct context.Context, sprintID uint64) (entity.Sprint, *errs.Error) {
	var sprint entity.Sprint
	txCtx := TransactionsContext{
		logger:             s.logger,
		transactionFactory: s.transactionFactory,
		stateSyncer:        s.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		taskIds, err := s.sprintTaskRelationDaoV2.FindTaskIDsBySprintIDWithTx(ct, tx, sprintID)
		if err != nil {
			return err
		}

		for _, taskId := range taskIds {
			_, err = s.removeTaskFromSprint(ct, tx, rtTx, sprintID, taskId)
			if err != nil {
				return err
			}
		}

		participantUserIDs, err := s.sprintParticipantDaoV2.FindParticipantIDsBySprintIDWithTx(ct, tx, sprintID)
		if err != nil {
			return err
		}

		sprint, err = s.sprintDaoV2.FindSprintByIDWithTx(ct, tx, sprintID)
		if err != nil {
			return err
		}

		for _, participantUserID := range participantUserIDs {
			deleteSprintParticipantMutation := mutation.NewDeleteSprintParticipant(
				s.logger,
				s.stateSyncer,
				s.sprintParticipantDao,
				s.sprintParticipantDaoV2,
				s.sprintDao,
				s.sprintDaoV2,
				participantUserID,
				sprintID)
			rtTx.AppendMutation(deleteSprintParticipantMutation)
			err = deleteSprintParticipantMutation.ExecuteV2(ct, tx)
			if err != nil {
				return err
			}

			// we need to prepare notifier in advance since sprint will be actually deleted later
			deleteSprintParticipantMutation.PrepareClientNotifiers(ct, tx)
		}

		return s.sprintDaoV2.DeleteSprint(ct, tx, sprintID)
	})

	if err != nil {
		return entity.Sprint{}, err
	}

	return sprint, nil
}

func (s Sprint) AddTaskToSprint(ct context.Context, sprintID uint64, taskID uint64) (entity.Task, *errs.Error) {
	var task entity.Task
	txCtx := TransactionsContext{
		logger:             s.logger,
		transactionFactory: s.transactionFactory,
		stateSyncer:        s.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		task, err = s.taskDaoV2.FindTaskByIDWithTx(ct, tx, taskID)
		if err != nil {
			return err
		}

		sprint, err := s.sprintDaoV2.FindSprintByIDWithTx(ct, tx, sprintID)
		if err != nil {
			return err
		}

		if sprint.OwningTeamID != task.OwningTeamID {
			return errs.NewError(errs.InvalidArgument, fmt.Sprintf("sprint and task must belong to the same team: sprintID=%v, taskID=%v", sprintID, taskID))
		}

		relation := entity.SprintTaskRelation{
			SprintID:  sprintID,
			TaskID:    taskID,
			CreatedAt: time.Now().UTC(),
		}
		createSprintTaskRelationMutation := mutation.NewCreateSprintTaskRelation(
			s.logger,
			s.stateSyncer,
			s.sprintTaskRelationDao,
			s.sprintTaskRelationDaoV2,
			s.sprintDao,
			s.sprintDaoV2,
			relation)
		rtTx.AppendMutation(createSprintTaskRelationMutation)
		err = createSprintTaskRelationMutation.ExecuteV2(ct, tx)
		if err != nil {
			return err
		}

		if !task.IsPlanned {
			task.IsPlanned = true
			updateTaskMutation := mutation.NewUpdateTask(
				s.logger,
				s.stateSyncer,
				s.taskDao,
				s.taskDaoV2,
				task)
			rtTx.AppendMutation(updateTaskMutation)
			err = updateTaskMutation.ExecuteV2(ct, tx)
			if err != nil {
				return err
			}
		}

		err = s.tryReduceBandwidth(ct, tx, rtTx, sprintID, task)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (s Sprint) CopyTasksToSprint(ct context.Context, toSprintID uint64, taskIDs []uint64) ([]entity.Task, *errs.Error) {
	if s.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		sprint, err := s.sprintDaoV2.FindSprintByID(ct, toSprintID)
		if err != nil {
			return nil, err
		}

		query := authorization.NewCloneTaskInTeamQuery(userID, sprint.OwningTeamID)
		hasPermission, err := s.authorizer.hasPermission(ct, query)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			return []entity.Task{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var tasks []entity.Task
	var newTaskIDs []uint64
	var newThreadIDs []uint64
	// TODO(yuhang): these genID requests should be batched in a single RPC
	for range taskIDs {
		genTaskIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "taskID"}
		genTaskIDRes, rpcErr := s.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genTaskIDReq)
		if rpcErr != nil {
			internalErr := errs.FromGRPCErr(rpcErr)
			return nil, internalErr
		}

		genThreadIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "threadID"}
		genThreadIDRes, rpcErr := s.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genThreadIDReq)
		if rpcErr != nil {
			internalErr := errs.FromGRPCErr(rpcErr)
			return nil, internalErr
		}

		newTaskIDs = append(newTaskIDs, genTaskIDRes.UniqueNumber)
		newThreadIDs = append(newThreadIDs, genThreadIDRes.UniqueNumber)
	}

	var err *errs.Error
	txCtx := TransactionsContext{
		logger:             s.logger,
		transactionFactory: s.transactionFactory,
		stateSyncer:        s.stateSyncer,
		ct:                 ct,
	}
	err = txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		for idx, taskID := range taskIDs {
			var task entity.Task
			task, err = s.copyTaskToSprint(ct, tx, rtTx, toSprintID, taskID, newTaskIDs[idx], newThreadIDs[idx])
			if err != nil {
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
		logger:             s.logger,
		transactionFactory: s.transactionFactory,
		stateSyncer:        s.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		for _, taskID := range taskIDs {
			task, err := s.moveTaskToSprint(ct, tx, rtTx, fromSprintID, toSprintID, taskID)

			if err != nil {
				continue
			}

			tasks = append(tasks, task)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (s Sprint) RemoveTaskFromSprint(ct context.Context, sprintID uint64, taskID uint64) (entity.Task, *errs.Error) {
	var task entity.Task
	txCtx := TransactionsContext{
		logger:             s.logger,
		transactionFactory: s.transactionFactory,
		stateSyncer:        s.stateSyncer,
		ct:                 ct,
	}
	err := txCtx.withTransactions(false, func(tx *transaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		var err *errs.Error
		task, err = s.removeTaskFromSprint(ct, tx, rtTx, sprintID, taskID)
		return err
	})

	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (s Sprint) copyTaskToSprint(
	ct context.Context,
	tx *transaction.Transaction,
	rtTx *realtime.Transaction,
	toSprintID uint64,
	taskID uint64,
	newTaskID uint64,
	newThreadID uint64,
) (entity.Task, *errs.Error) {
	task, err := s.taskDaoV2.FindTaskByIDWithTx(ct, tx, taskID)
	if err != nil {
		return entity.Task{}, err
	}

	sprint, err := s.sprintDaoV2.FindSprintByIDWithTx(ct, tx, toSprintID)
	if err != nil {
		return entity.Task{}, err
	}

	if sprint.OwningTeamID != task.OwningTeamID {
		return entity.Task{}, errs.NewError(
			errs.InvalidArgument,
			fmt.Sprintf("sprint and task must belong to the same team: sprintID=%v, taskID=%v", toSprintID, newTaskID))
	}

	err = s.threadDaoV2.CreateThread(ct, tx, newThreadID)
	if err != nil {
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

	createTaskMutation := mutation.NewCreateTask(
		s.logger,
		s.stateSyncer,
		s.taskDao,
		s.taskDaoV2,
		newTask,
	)
	err = createTaskMutation.ExecuteV2(ct, tx)
	if err != nil {
		return entity.Task{}, err
	}

	rtTx.AppendMutation(createTaskMutation)
	relation := entity.SprintTaskRelation{
		SprintID:  toSprintID,
		TaskID:    newTaskID,
		CreatedAt: time.Now().UTC(),
	}
	createSprintTaskRelationMutation := mutation.NewCreateSprintTaskRelation(
		s.logger,
		s.stateSyncer,
		s.sprintTaskRelationDao,
		s.sprintTaskRelationDaoV2,
		s.sprintDao,
		s.sprintDaoV2,
		relation)
	rtTx.AppendMutation(createSprintTaskRelationMutation)
	err = createSprintTaskRelationMutation.ExecuteV2(ct, tx)
	if err != nil {
		return entity.Task{}, err
	}

	if !task.IsPlanned {
		task.IsPlanned = true
		updateTaskMutation := mutation.NewUpdateTask(
			s.logger,
			s.stateSyncer,
			s.taskDao,
			s.taskDaoV2,
			task)
		rtTx.AppendMutation(updateTaskMutation)
		err = updateTaskMutation.ExecuteV2(ct, tx)
		if err != nil {
			return entity.Task{}, err
		}
	}

	err = s.tryReduceBandwidth(ct, tx, rtTx, toSprintID, task)
	if err != nil {
		return entity.Task{}, err
	}

	return newTask, nil
}

func (s Sprint) moveTaskToSprint(
	ct context.Context,
	tx *transaction.Transaction,
	rtTx *realtime.Transaction,
	fromSprintID uint64,
	toSprintID uint64,
	taskID uint64,
) (entity.Task, *errs.Error) {
	sprintIDs, err := s.sprintTaskRelationDaoV2.FindSprintIDsByTaskIDWithTx(ct, tx, taskID)
	if err != nil {
		return entity.Task{}, err
	}

	foundSprintIDs := collect.Filter(sprintIDs, func(currSprintID uint64) bool {
		return currSprintID == fromSprintID
	})
	if len(foundSprintIDs) < 1 {
		return entity.Task{}, errs.NewError(
			errs.NotFound,
			fmt.Sprintf("relation not found: sprintID=%v, taskID=%v", fromSprintID, taskID))
	}

	err = s.sprintTaskRelationDaoV2.DeleteSprintTaskRelation(ct, tx, fromSprintID, taskID)
	if err != nil {
		return entity.Task{}, err
	}

	relation := entity.SprintTaskRelation{
		SprintID:  toSprintID,
		TaskID:    taskID,
		CreatedAt: time.Now().UTC(),
	}

	err = s.sprintTaskRelationDaoV2.CreateSprintTaskRelation(ct, tx, relation)
	if err != nil {
		return entity.Task{}, err
	}

	task, err := s.taskDaoV2.FindTaskByIDWithTx(ct, tx, taskID)
	if err != nil {
		return entity.Task{}, err
	}

	err = s.tryIncreaseBandwidth(ct, tx, rtTx, fromSprintID, task)
	if err != nil {
		return entity.Task{}, err
	}

	err = s.tryReduceBandwidth(ct, tx, rtTx, toSprintID, task)
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (s Sprint) removeTaskFromSprint(ct context.Context, tx *transaction.Transaction, rtTx *realtime.Transaction, sprintID uint64, taskID uint64) (entity.Task, *errs.Error) {
	sprintIDs, err := s.sprintTaskRelationDaoV2.FindSprintIDsByTaskIDWithTx(ct, tx, taskID)
	if err != nil {
		return entity.Task{}, err
	}

	foundSprintIDs := collect.Filter(sprintIDs, func(currSprintID uint64) bool {
		return currSprintID == sprintID
	})
	if len(foundSprintIDs) < 1 {
		return entity.Task{}, errs.NewError(
			errs.NotFound,
			fmt.Sprintf("relation not found: sprintID=%v, taskID=%v", sprintID, taskID))
	}

	task, err := s.taskDaoV2.FindTaskByIDWithTx(ct, tx, taskID)
	if err != nil {
		return entity.Task{}, err
	}

	deleteSprintTaskRelationMutation := mutation.NewDeleteSprintTaskRelation(
		s.logger,
		s.stateSyncer,
		s.sprintTaskRelationDao,
		s.sprintTaskRelationDaoV2,
		sprintID,
		task,
	)
	rtTx.AppendMutation(deleteSprintTaskRelationMutation)
	err = deleteSprintTaskRelationMutation.ExecuteV2(ct, tx)
	if err != nil {
		return entity.Task{}, err
	}

	//if there is no other sprint that the task can move to,  put it into backlog
	if len(sprintIDs) <= 1 {
		task.IsPlanned = false
		updateTaskMutation := mutation.NewUpdateTask(
			s.logger,
			s.stateSyncer,
			s.taskDao,
			s.taskDaoV2,
			task)
		rtTx.AppendMutation(updateTaskMutation)
		err = updateTaskMutation.ExecuteV2(ct, tx)
		if err != nil {
			return entity.Task{}, err
		}
	}

	err = s.tryIncreaseBandwidth(ct, tx, rtTx, sprintID, task)
	if err != nil {
		return entity.Task{}, err
	}

	return task, nil
}

func (s Sprint) tryReduceBandwidth(
	ct context.Context,
	tx *transaction.Transaction,
	rtTx *realtime.Transaction,
	sprintID uint64,
	task entity.Task,
) *errs.Error {
	if task.OwnerUserID != nil && task.Effort != nil {
		newSprintParticipant, err := s.sprintParticipantDaoV2.FindParticipantWithTx(ct, tx, sprintID, *task.OwnerUserID)
		if err != nil {
			// TODO: this should be removed once the sprint participants are backfilled
			if err.Code == errs.NotFound {
				return nil
			}

			return err
		}

		newSprintParticipant.UnusedBandwidth -= *task.Effort
		updateSprintParticipantMutation := mutation.NewUpdateSprintParticipant(
			s.logger,
			s.stateSyncer,
			s.sprintParticipantDao,
			s.sprintParticipantDaoV2,
			s.sprintDao,
			s.sprintDaoV2,
			newSprintParticipant)
		rtTx.AppendMutation(updateSprintParticipantMutation)
		err = updateSprintParticipantMutation.ExecuteV2(ct, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s Sprint) tryIncreaseBandwidth(
	ct context.Context,
	tx *transaction.Transaction,
	rtTx *realtime.Transaction,
	sprintID uint64,
	task entity.Task,
) *errs.Error {
	if task.OwnerUserID != nil && task.Effort != nil {
		oldSprintParticipant, err := s.sprintParticipantDaoV2.FindParticipantWithTx(ct, tx, sprintID, *task.OwnerUserID)
		if err != nil {
			// TODO: this should be removed once the sprint participants are backfilled
			if err.Code == errs.NotFound {
				return nil
			}

			return err
		}

		oldSprintParticipant.UnusedBandwidth += *task.Effort
		updateSprintParticipantMutation := mutation.NewUpdateSprintParticipant(
			s.logger,
			s.stateSyncer,
			s.sprintParticipantDao,
			s.sprintParticipantDaoV2,
			s.sprintDao,
			s.sprintDaoV2,
			oldSprintParticipant)
		rtTx.AppendMutation(updateSprintParticipantMutation)
		err = updateSprintParticipantMutation.ExecuteV2(ct, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewSprint(
	logger telemetry.Logger,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	stateSyncer *realtime.StateSyncer,
	authorizer Authorizer,
	featureToggles feature.Toggles,
	transactionFactory transaction.Factory,
	taskDao dao.Task,
	taskDaoV2 daov2.Task,
	sprintDao dao.Sprint,
	sprintDaoV2 daov2.Sprint,
	teamDaoV2 daov2.Team,
	sprintTaskRelationDao dao.SprintTaskRelation,
	sprintTaskRelationDaoV2 daov2.SprintTaskRelation,
	sprintParticipantDao dao.SprintParticipant,
	sprintParticipantDaoV2 daov2.SprintParticipant,
	teamMemberDao dao.TeamMember,
	teamMemberDaoV2 daov2.TeamMember,
	threadDaoV2 daov2.Thread,
) Sprint {
	return Sprint{
		logger:                  logger,
		cloudClientRegistry:     cloudClientRegistry,
		stateSyncer:             stateSyncer,
		authorizer:              authorizer,
		featureToggles:          featureToggles,
		transactionFactory:      transactionFactory,
		taskDao:                 taskDao,
		taskDaoV2:               taskDaoV2,
		sprintDao:               sprintDao,
		sprintDaoV2:             sprintDaoV2,
		teamDaoV2:               teamDaoV2,
		sprintTaskRelationDao:   sprintTaskRelationDao,
		sprintTaskRelationDaoV2: sprintTaskRelationDaoV2,
		sprintParticipantDao:    sprintParticipantDao,
		sprintParticipantDaoV2:  sprintParticipantDaoV2,
		teamMemberDao:           teamMemberDao,
		teamMemberDaoV2:         teamMemberDaoV2,
		threadDaoV2:             threadDaoV2,
	}
}
