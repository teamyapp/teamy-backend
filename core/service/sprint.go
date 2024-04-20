package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/teamyapp/teamy-backend/core/cache"

	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/app/client"
	cloudAuthorization "github.com/teamyapp/cloud/libs/authorization"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/ctx"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/authorization"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/transaction"
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
	transactionGroupFactory transaction.GroupFactory
	cloudClientRegistry     *client.Registry
	stateSyncer             *realtime.StateSyncer
	authorizer              client.Authorizer
	featureToggles          feature.Toggles
	transactionFactory      cloudTransaction.Factory
	cache                   *cache.TimeBasedCache[string, any]
	taskDao                 dao.Task
	sprintDao               dao.Sprint
	teamDao                 dao.Team
	sprintTaskRelationDao   dao.SprintTaskRelation
	sprintParticipantDao    dao.SprintParticipant
	teamMemberDao           dao.TeamMember
	threadDao               dao.Thread
	db                      *sql.DB
}

func (s Sprint) FindSprintsInTeam(ct context.Context, teamID uint64, filter *SprintFilter) ([]entity.Sprint, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if s.featureToggles.EnableAuthorization {
		if !ok {
			return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewReadInTeamQuery(userID, teamID)
		hasPermission, err := s.authorizer.HasPermission(ct, query)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			return nil, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	sprints, err := s.sprintDao.FindSprintsByTeamID(ct, teamID)
	if err != nil {
		return nil, err
	}

	if s.featureToggles.EnableAuthorization {
		authorizedSprints, err := client.FilterAuthorizedItems(
			ct,
			s.authorizer,
			sprints,
			func(sprint entity.Sprint) cloudAuthorization.Query {
				return authorization.NewReadInSprintQuery(userID, sprint.ID)
			})
		if err != nil {
			return nil, err
		}

		sprints = authorizedSprints
	}

	if filter != nil {
		sprints = filterSprints(sprints, *filter)
	}

	return sprints, nil
}

func (s Sprint) FindParticipantsInSprint(ct context.Context, sprintID uint64) ([]entity.SprintParticipant, *errs.Error) {
	if s.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewReadInSprintQuery(userID, sprintID)
		hasPermission, err := s.authorizer.HasPermission(ct, query)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			return nil, errs.NewError(
				errs.PermissionDenied,
				fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	if s.featureToggles.EnableCache {
		value, cacheErr := s.cache.Get(ct, findParticipantsInSprintCacheKey(sprintID))
		if cacheErr == nil {
			return value.([]entity.SprintParticipant), nil
		}

		var cacheKeyNotFoundErr cache.KeyNotFoundErr[string]
		if !errors.As(cacheErr, &cacheKeyNotFoundErr) {
			return nil, errs.NewError(errs.Unknown, cacheErr.Error())
		}
	}

	participants, err := s.sprintParticipantDao.FindParticipantsBySprintID(ct, sprintID)
	if err != nil {
		return nil, err
	}

	if s.featureToggles.EnableCache {
		cacheErr := s.cache.SetIfExpired(ct, findParticipantsInSprintCacheKey(sprintID), participants)
		if cacheErr != nil {
			return nil, errs.NewError(errs.Unknown, cacheErr.Error())
		}
	}

	return participants, nil
}

func (s Sprint) FindSprints(ct context.Context, filter *SprintFilter) ([]entity.Sprint, *errs.Error) {
	sprints, err := s.sprintDao.FindAllSprints(ct)
	if err != nil {
		return nil, err
	}

	if s.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		authorizedSprints, err := client.FilterAuthorizedItems(
			ct,
			s.authorizer,
			sprints,
			func(sprint entity.Sprint) cloudAuthorization.Query {
				return authorization.NewReadInSprintQuery(userID, sprint.ID)
			})
		if err != nil {
			return nil, err
		}

		sprints = authorizedSprints
	}

	if filter != nil {
		sprints = filterSprints(sprints, *filter)
	}

	return sprints, nil
}

func (s Sprint) GetActiveSprint(ct context.Context, teamID uint64) (*entity.Sprint, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if s.featureToggles.EnableAuthorization {
		if !ok {
			return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewReadInTeamQuery(userID, teamID)
		hasPermission, err := s.authorizer.HasPermission(ct, query)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			return nil, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var sprint *entity.Sprint
	err := s.transactionGroupFactory.WithTransactionGroup(
		ct,
		true,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			team, err := s.teamDao.FindTeamByIDWithTx(ct, tx, teamID)
			if err != nil {
				return err
			}

			if team.ActiveSprintID == nil {
				return nil
			}

			sprintRes, err := s.sprintDao.FindSprintByID(ct, *team.ActiveSprintID)
			if err != nil {
				return err
			}

			sprint = &sprintRes
			return nil
		})
	if err != nil {
		return nil, err
	}

	if sprint == nil {
		return nil, errs.NewError(errs.NotFound, "active sprint not found")
	}

	if s.featureToggles.EnableAuthorization {
		if !ok {
			return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewReadInSprintQuery(userID, sprint.ID)
		hasPermission, err := s.authorizer.HasPermission(ct, query)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			return nil, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
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
		hasPermission, err := s.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Team{}, err
		}

		if !hasPermission {
			return entity.Team{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var team entity.Team
	err := s.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			team, err = s.teamDao.FindTeamByIDWithTx(ct, tx, teamID)
			if err != nil {
				return err
			}

			sprint, err := s.sprintDao.FindSprintByIDWithTx(ct, tx, sprintID)
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
				team,
			)
			rtTx.AppendMutation(updateTeamMutation)
			return updateTeamMutation.Execute(ct, tx)
		})

	if err != nil {
		return entity.Team{}, err
	}

	return team, nil
}

func (s Sprint) FindCurrentAndFutureSprints(ct context.Context, teamID uint64) ([]entity.Sprint, *errs.Error) {
	userID, ok := ctx.UserIDFromContext(ct)
	if s.featureToggles.EnableAuthorization {
		if !ok {
			return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewReadInTeamQuery(userID, teamID)
		hasPermission, err := s.authorizer.HasPermission(ct, query)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			return nil, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var sprints []entity.Sprint
	err := s.transactionGroupFactory.WithTransactionGroup(
		ct,
		true,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			sprints, err = s.sprintDao.FindSprintsByTeamIDWithTx(ct, tx, teamID)
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

	if s.featureToggles.EnableAuthorization {
		authorizedSprints, err := client.FilterAuthorizedItems(
			ct,
			s.authorizer,
			sprints,
			func(sprint entity.Sprint) cloudAuthorization.Query {
				return authorization.NewReadInSprintQuery(userID, sprint.ID)
			})
		if err != nil {
			return nil, err
		}

		sprints = authorizedSprints
	}

	return sprints, nil
}

func (s Sprint) FindSprintByID(ct context.Context, sprintID uint64) (entity.Sprint, *errs.Error) {
	if s.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.Sprint{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewReadInSprintQuery(userID, sprintID)
		hasPermission, err := s.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Sprint{}, err
		}

		if !hasPermission {
			return entity.Sprint{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	if s.featureToggles.EnableCache {
		value, cacheErr := s.cache.Get(ct, findSprintByIDCacheKey(sprintID))
		if cacheErr == nil {
			return value.(entity.Sprint), nil
		}

		var cacheKeyNotFoundErr cache.KeyNotFoundErr[string]
		if !errors.As(cacheErr, &cacheKeyNotFoundErr) {
			return entity.Sprint{}, errs.NewError(errs.Unknown, cacheErr.Error())
		}
	}

	var sprint entity.Sprint
	err := s.transactionGroupFactory.WithTransactionGroup(
		ct,
		true,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			sprint, err = s.sprintDao.FindSprintByIDWithTx(ct, tx, sprintID)
			return err
		})

	if err != nil {
		return entity.Sprint{}, err
	}

	if s.featureToggles.EnableCache {
		cacheErr := s.cache.SetIfExpired(ct, findSprintByIDCacheKey(sprintID), sprint)
		if cacheErr != nil {
			return entity.Sprint{}, errs.NewError(errs.Unknown, cacheErr.Error())
		}
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
		hasPermission, err := s.authorizer.HasPermission(ct, query)
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
	err := s.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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
				s.sprintDao,
				sprint)

			rtTx.AppendMutation(createSprintMutation)
			err := createSprintMutation.Execute(ct, tx)
			if err != nil {
				return err
			}

			teamMembers, err := s.teamMemberDao.FindTeamMembersByTeamIDWithTx(ct, tx, teamID)
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
					s.sprintDao,
					participant)
				rtTx.AppendMutation(createSprintParticipantMutation)
				err = createSprintParticipantMutation.Execute(ct, tx)
				if err != nil {
					return err
				}
			}

			return nil
		})

	if err != nil {
		return entity.Sprint{}, err
	}

	// TODO(magicoder10): if failed to register/assign resource, there will be inconsistent state. Cross-service transaction
	// protection will be covered in stage 2
	if s.featureToggles.EnableAuthorization {
		err = s.authorizer.RegisterResource(ct, authorization.SprintResourceType, sprint.ID)
		if err != nil {
			return entity.Sprint{}, err
		}

		err = s.authorizer.AssignParentResource(ct, authorization.SprintResourceType, sprint.ID, authorization.TeamResourceType, sprint.OwningTeamID)
		if err != nil {
			return entity.Sprint{}, err
		}
	}

	return sprint, nil
}

func (s Sprint) DeleteSprint(ct context.Context, sprintID uint64) (entity.Sprint, *errs.Error) {
	if s.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.Sprint{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewDeleteInSprintQuery(userID, sprintID)
		hasPermission, err := s.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Sprint{}, err
		}

		if !hasPermission {
			return entity.Sprint{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var sprint entity.Sprint
	err := s.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			taskIds, err := s.sprintTaskRelationDao.FindTaskIDsBySprintIDWithTx(ct, tx, sprintID)
			if err != nil {
				return err
			}

			for _, taskId := range taskIds {
				_, err = s.removeTaskFromSprint(ct, tx, rtTx, sprintID, taskId, false)
				if err != nil {
					return err
				}
			}

			participantUserIDs, err := s.sprintParticipantDao.FindParticipantIDsBySprintIDWithTx(ct, tx, sprintID)
			if err != nil {
				return err
			}

			sprint, err = s.sprintDao.FindSprintByIDWithTx(ct, tx, sprintID)
			if err != nil {
				return err
			}

			for _, participantUserID := range participantUserIDs {
				deleteSprintParticipantMutation := mutation.NewDeleteSprintParticipant(
					s.logger,
					s.stateSyncer,
					s.sprintParticipantDao,
					s.sprintDao,
					participantUserID,
					sprintID)
				rtTx.AppendMutation(deleteSprintParticipantMutation)
				err = deleteSprintParticipantMutation.Execute(ct, tx)
				if err != nil {
					return err
				}

				// we need to prepare notifier in advance since sprint will be actually deleted later
				deleteSprintParticipantMutation.PrepareClientNotifiers(ct, tx)
			}

			deleteSprintMutation := mutation.NewDeleteSprint(
				s.logger,
				s.stateSyncer,
				s.sprintDao,
				sprint,
			)

			rtTx.AppendMutation(deleteSprintMutation)
			return deleteSprintMutation.Execute(ct, tx)
		})

	if err != nil {
		return entity.Sprint{}, err
	}

	// TODO: clean up resource relations in authorization service
	return sprint, nil
}

func (s Sprint) AddTaskToSprint(ct context.Context, sprintID uint64, taskID uint64) (entity.Task, *errs.Error) {
	if s.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.Task{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewAddTaskToInSprintQuery(userID, sprintID)
		hasPermission, err := s.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Task{}, err
		}

		if !hasPermission {
			return entity.Task{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var task entity.Task
	err := s.transactionGroupFactory.WithTransactionGroup(
		ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			task, err = s.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
			if err != nil {
				return err
			}

			sprint, err := s.sprintDao.FindSprintByIDWithTx(ct, tx, sprintID)
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
				s.sprintDao,
				relation)
			rtTx.AppendMutation(createSprintTaskRelationMutation)
			err = createSprintTaskRelationMutation.Execute(ct, tx)
			if err != nil {
				return err
			}

			if !task.IsScheduled {
				task.IsScheduled = true
				updateTaskMutation := mutation.NewUpdateTask(
					s.logger,
					s.stateSyncer,
					s.taskDao,
					task)
				rtTx.AppendMutation(updateTaskMutation)
				err = updateTaskMutation.Execute(ct, tx)
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

	// TODO: update resource relations in authorization service
	return task, nil
}

func (s Sprint) CopyTasksToSprint(
	ct context.Context,
	toSprintID uint64,
	taskIDs []uint64,
) ([]entity.Task, *errs.Error) {
	if s.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		for _, taskID := range taskIDs {
			query := authorization.NewReadInTaskQuery(userID, taskID)
			hasPermission, err := s.authorizer.HasPermission(ct, query)
			if err != nil {
				return nil, err
			}

			if !hasPermission {
				return nil, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
			}
		}

		sprint, err := s.sprintDao.FindSprintByID(ct, toSprintID)
		if err != nil {
			return nil, err
		}

		query := authorization.NewCreateTaskInTeamQuery(userID, sprint.OwningTeamID)
		hasPermission, err := s.authorizer.HasPermission(ct, query)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			return nil, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}

		query = authorization.NewAddTaskToInSprintQuery(userID, toSprintID)
		hasPermission, err = s.authorizer.HasPermission(ct, query)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			return nil, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var tasks []entity.Task
	var newTaskIDs []uint64
	var newThreadIDs []uint64
	// TODO(magicoder10): these genID requests should be batched in a single RPC
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
	err = s.transactionGroupFactory.WithTransactionGroup(
		ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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

	// TODO: update resource relations in authorization service
	return tasks, nil
}

func (s Sprint) MoveTasksToSprint(
	ct context.Context,
	fromSprintID uint64,
	toSprintID uint64,
	taskIDs []uint64,
) ([]entity.Task, *errs.Error) {
	if s.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return nil, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewRemoveTaskFromInSprintQuery(userID, fromSprintID)
		hasPermission, err := s.authorizer.HasPermission(ct, query)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			return nil, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}

		query = authorization.NewAddTaskToInSprintQuery(userID, toSprintID)
		hasPermission, err = s.authorizer.HasPermission(ct, query)
		if err != nil {
			return nil, err
		}

		if !hasPermission {
			return nil, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var tasks []entity.Task
	err := s.transactionGroupFactory.WithTransactionGroup(
		ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
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

	// TODO: update resource relations in authorization service
	return tasks, nil
}

func (s Sprint) RemoveTaskFromSprint(ct context.Context, sprintID uint64, taskID uint64) (entity.Task, *errs.Error) {
	if s.featureToggles.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			return entity.Task{}, errs.NewError(errs.Unauthenticated, "user ID not found")
		}

		query := authorization.NewRemoveTaskFromInSprintQuery(userID, sprintID)
		hasPermission, err := s.authorizer.HasPermission(ct, query)
		if err != nil {
			return entity.Task{}, err
		}

		if !hasPermission {
			return entity.Task{}, errs.NewError(errs.PermissionDenied, fmt.Sprintf("permission denied: authorization query=%v", query))
		}
	}

	var task entity.Task
	err := s.transactionGroupFactory.WithTransactionGroup(
		ct, false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			var err *errs.Error
			task, err = s.removeTaskFromSprint(ct, tx, rtTx, sprintID, taskID, true)
			return err
		})

	if err != nil {
		return entity.Task{}, err
	}

	// TODO: update resource relations in authorization service
	return task, nil
}

func (s Sprint) copyTaskToSprint(
	ct context.Context,
	tx *cloudTransaction.Transaction,
	rtTx *realtime.Transaction,
	toSprintID uint64,
	taskID uint64,
	newTaskID uint64,
	newThreadID uint64,
) (entity.Task, *errs.Error) {
	task, err := s.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
	if err != nil {
		return entity.Task{}, err
	}

	sprint, err := s.sprintDao.FindSprintByIDWithTx(ct, tx, toSprintID)
	if err != nil {
		return entity.Task{}, err
	}

	if sprint.OwningTeamID != task.OwningTeamID {
		return entity.Task{}, errs.NewError(
			errs.InvalidArgument,
			fmt.Sprintf("sprint and task must belong to the same team: sprintID=%v, taskID=%v", toSprintID, newTaskID))
	}

	err = s.threadDao.CreateThread(ct, tx, newThreadID)
	if err != nil {
		return entity.Task{}, err
	}

	newTask := entity.Task{
		ID:               newTaskID,
		Goal:             task.Goal,
		Context:          task.Context,
		Status:           task.Status,
		IsScheduled:      task.IsScheduled,
		IsPlanned:        task.IsPlanned,
		CreatorUserID:    task.CreatorUserID,
		OwningTeamID:     task.OwningTeamID,
		Effort:           task.Effort,
		Priority:         task.Priority,
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
		newTask,
	)
	err = createTaskMutation.Execute(ct, tx)
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
		s.sprintDao,
		relation)
	rtTx.AppendMutation(createSprintTaskRelationMutation)
	err = createSprintTaskRelationMutation.Execute(ct, tx)
	if err != nil {
		return entity.Task{}, err
	}

	if !task.IsScheduled {
		task.IsScheduled = true
		updateTaskMutation := mutation.NewUpdateTask(
			s.logger,
			s.stateSyncer,
			s.taskDao,
			task)
		rtTx.AppendMutation(updateTaskMutation)
		err = updateTaskMutation.Execute(ct, tx)
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
	tx *cloudTransaction.Transaction,
	rtTx *realtime.Transaction,
	fromSprintID uint64,
	toSprintID uint64,
	taskID uint64,
) (entity.Task, *errs.Error) {
	sprintIDs, err := s.sprintTaskRelationDao.FindSprintIDsByTaskIDWithTx(ct, tx, taskID)
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

	err = s.sprintTaskRelationDao.DeleteSprintTaskRelation(ct, tx, fromSprintID, taskID)
	if err != nil {
		return entity.Task{}, err
	}

	relation := entity.SprintTaskRelation{
		SprintID:  toSprintID,
		TaskID:    taskID,
		CreatedAt: time.Now().UTC(),
	}

	err = s.sprintTaskRelationDao.CreateSprintTaskRelation(ct, tx, relation)
	if err != nil {
		return entity.Task{}, err
	}

	task, err := s.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
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

func (s Sprint) removeTaskFromSprint(
	ct context.Context,
	tx *cloudTransaction.Transaction,
	rtTx *realtime.Transaction,
	sprintID uint64,
	taskID uint64,
	adjustBandwidth bool,
) (entity.Task, *errs.Error) {
	sprintIDs, err := s.sprintTaskRelationDao.FindSprintIDsByTaskIDWithTx(ct, tx, taskID)
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

	task, err := s.taskDao.FindTaskByIDWithTx(ct, tx, taskID)
	if err != nil {
		return entity.Task{}, err
	}

	deleteSprintTaskRelationMutation := mutation.NewDeleteSprintTaskRelation(
		s.logger,
		s.stateSyncer,
		s.sprintTaskRelationDao,
		sprintID,
		task,
	)
	rtTx.AppendMutation(deleteSprintTaskRelationMutation)
	err = deleteSprintTaskRelationMutation.Execute(ct, tx)
	if err != nil {
		return entity.Task{}, err
	}

	//if there is no other sprint that the task can move to,  put it into backlog
	if len(sprintIDs) <= 1 {
		task.IsScheduled = false
		updateTaskMutation := mutation.NewUpdateTask(
			s.logger,
			s.stateSyncer,
			s.taskDao,
			task)
		rtTx.AppendMutation(updateTaskMutation)
		err = updateTaskMutation.Execute(ct, tx)
		if err != nil {
			return entity.Task{}, err
		}
	}

	if adjustBandwidth {
		err = s.tryIncreaseBandwidth(ct, tx, rtTx, sprintID, task)
		if err != nil {
			return entity.Task{}, err
		}
	}

	return task, nil
}

func (s Sprint) tryReduceBandwidth(
	ct context.Context,
	tx *cloudTransaction.Transaction,
	rtTx *realtime.Transaction,
	sprintID uint64,
	task entity.Task,
) *errs.Error {
	if task.OwnerUserID != nil && task.Effort != nil {
		newSprintParticipant, err := s.sprintParticipantDao.FindParticipantWithTx(ct, tx, sprintID, *task.OwnerUserID)
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
			s.sprintDao,
			newSprintParticipant)
		rtTx.AppendMutation(updateSprintParticipantMutation)
		err = updateSprintParticipantMutation.Execute(ct, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func (s Sprint) tryIncreaseBandwidth(
	ct context.Context,
	tx *cloudTransaction.Transaction,
	rtTx *realtime.Transaction,
	sprintID uint64,
	task entity.Task,
) *errs.Error {
	if task.OwnerUserID != nil && task.Effort != nil {
		oldSprintParticipant, err := s.sprintParticipantDao.FindParticipantWithTx(ct, tx, sprintID, *task.OwnerUserID)
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
			s.sprintDao,
			oldSprintParticipant)
		rtTx.AppendMutation(updateSprintParticipantMutation)

		err = updateSprintParticipantMutation.Execute(ct, tx)
		if err != nil {
			return err
		}
	}

	return nil
}

func NewSprint(
	logger telemetry.Logger,
	transactionGroupFactory transaction.GroupFactory,
	cloudClientRegistry *client.Registry,
	stateSyncer *realtime.StateSyncer,
	authorizer client.Authorizer,
	featureToggles feature.Toggles,
	transactionFactory cloudTransaction.Factory,
	cache *cache.TimeBasedCache[string, any],
	taskDao dao.Task,
	sprintDao dao.Sprint,
	teamDao dao.Team,
	sprintTaskRelationDao dao.SprintTaskRelation,
	sprintParticipantDao dao.SprintParticipant,
	teamMemberDao dao.TeamMember,
	threadDao dao.Thread,
) Sprint {
	return Sprint{
		logger:                  logger,
		transactionGroupFactory: transactionGroupFactory,
		cloudClientRegistry:     cloudClientRegistry,
		stateSyncer:             stateSyncer,
		authorizer:              authorizer,
		featureToggles:          featureToggles,
		transactionFactory:      transactionFactory,
		cache:                   cache,
		taskDao:                 taskDao,
		sprintDao:               sprintDao,
		teamDao:                 teamDao,
		sprintTaskRelationDao:   sprintTaskRelationDao,
		sprintParticipantDao:    sprintParticipantDao,
		teamMemberDao:           teamMemberDao,
		threadDao:               threadDao,
	}
}

func findParticipantsInSprintCacheKey(sprintID uint64) string {
	return fmt.Sprintf("FindParticipantsInSprint(%v)", sprintID)
}

func findSprintByIDCacheKey(sprintID uint64) string {
	return fmt.Sprintf("FindSprintByID(%v)", sprintID)
}
