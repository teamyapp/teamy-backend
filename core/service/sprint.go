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
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
	"github.com/teamyapp/teamy-backend/core/feature"
	"github.com/teamyapp/teamy-backend/core/mutation"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

const timePerWeek = 7 * 24 * time.Hour

type CreateSprintInput struct {
	StartAt time.Time
	EndAt   time.Time
}

type Sprint struct {
	dataCollector         obs.DataCollector
	cloudClientRegistry   *cloudAPI.ClientRegistry
	stateSyncer           *realtime.StateSyncer
	authorizer            Authorizer
	taskDao               dao.Task
	sprintDao             dao.Sprint
	sprintTaskRelationDao dao.SprintTaskRelation
	sprintParticipantDao  dao.SprintParticipant
	teamMemberDao         dao.TeamMember
	taskService           Task
}

func (s Sprint) FindTasksInSprint(
	ct context.Context,
	sprintID uint64,
	filter *TaskFilter,
) ([]entity.Task, error) {
	taskIDs, err := s.sprintTaskRelationDao.FindTaskIDsBySprintID(ct, sprintID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	tasks, err := s.taskDao.FindTasksByIDs(ct, taskIDs)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	if filter != nil {
		tasks = filterTasks(tasks, *filter)
	}

	return tasks, nil
}

func (s Sprint) FindParticipantsInSprint(ct context.Context, sprintID uint64) ([]entity.SprintParticipant, error) {
	participants, err := s.sprintParticipantDao.FindParticipantsBySprintID(ct, sprintID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	return participants, nil
}

func (s Sprint) FindSprints(ct context.Context, filter *SprintFilter) ([]entity.Sprint, error) {
	sprints, err := s.sprintDao.FindAllSprints(ct)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	if filter != nil {
		sprints = filterSprints(sprints, *filter)
	}

	return sprints, nil
}

func (s Sprint) FindCurrentSprint(ct context.Context, teamID uint64) (entity.Sprint, error) {
	sprints, err := s.sprintDao.FindSprintsByTeamID(ct, teamID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Sprint{}, err
	}

	now := time.Now().UTC()
	sprints = collect.Filter(sprints, func(sprint entity.Sprint) bool {
		if now.Before(sprint.StartAt.UTC()) || now.After(sprint.EndAt.UTC()) {
			return false
		}

		return true
	})
	if len(sprints) < 1 {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"TeamID":      teamID,
				"CurrentTime": now.UTC(),
			},
		})
		return entity.Sprint{}, err
	}

	if len(sprints) > 1 {
		err = errors.New("team has more than one sprint")
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"TeamID": teamID,
			},
		})
		return entity.Sprint{}, err
	}

	return sprints[0], nil
}

func (s Sprint) FindCurrentAndFutureSprints(ct context.Context, teamID uint64) ([]entity.Sprint, error) {
	sprints, err := s.sprintDao.FindSprintsByTeamID(ct, teamID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	now := time.Now().UTC()
	return collect.Filter(sprints, func(sprint entity.Sprint) bool {
		if sprint.EndAt.UTC().Before(now) {
			return false
		}

		return true
	}), nil
}

func (s Sprint) FindSprintByID(ct context.Context, sprintID uint64) (entity.Sprint, error) {
	return s.sprintDao.FindSprintByID(ct, sprintID)
}

func (s Sprint) CreateSprint(ct context.Context, teamID uint64, sprint CreateSprintInput) (entity.Sprint, error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			err := errors.New("user id not found")
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Sprint{}, err
		}

		query := authorization.NewCreateSprintQuery(userID, teamID)
		hasPermission, err := s.authorizer.hasPermission(ct, query)
		if err != nil {
			s.authorizer.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Sprint{}, err
		}

		if !hasPermission {
			return entity.Sprint{}, authorization.Error{
				Code:    authorization.UnauthorizedErrorCode,
				Message: fmt.Sprintf("Unauthorized: %v", query),
			}
		}
	}

	genSprintIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "sprintID"}
	genSprintIDRes, err := s.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genSprintIDReq)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Sprint{}, err
	}

	sp := entity.Sprint{
		ID:           genSprintIDRes.UniqueNumber,
		StartAt:      sprint.StartAt.UTC(),
		EndAt:        sprint.EndAt.UTC(),
		CreatedAt:    time.Now().UTC(),
		OwningTeamID: teamID,
	}
	err = s.sprintDao.CreateSprint(ct, sp)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Sprint{}, err
	}

	if feature.EnableAuthorization {
		err = s.authorizer.registerResource(ct, authorization.SprintResourceType, sp.ID)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Sprint{}, err
		}

		err = s.authorizer.assignParentResource(ct, authorization.SprintResourceType, sp.ID, authorization.TeamResourceType, sp.OwningTeamID)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Sprint{}, err
		}
	}

	teamMembers, err := s.teamMemberDao.FindTeamMembersByTeamID(ct, teamID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Sprint{}, err
	}

	sprintLength := sprint.EndAt.UTC().Sub(sprint.StartAt.UTC())
	numOfWeeks := sprintLength / timePerWeek
	// TODO: fetch from team settings

	realTimeTransaction := realtime.NewTransaction(s.dataCollector, s.stateSyncer)
	for _, teamMember := range teamMembers {
		totalBandwidth := teamMember.WeeklyBandwidth * numOfWeeks
		participant := entity.SprintParticipant{
			SprintID:        sp.ID,
			UserID:          teamMember.UserID,
			TotalBandwidth:  totalBandwidth,
			UnusedBandwidth: totalBandwidth,
			CreatedAt:       time.Now(),
		}
		createSprintParticipantMutation := mutation.NewCreateSprintParticipantMutation(
			s.dataCollector,
			s.stateSyncer,
			s.sprintParticipantDao,
			s.sprintDao,
			participant)
		err = realTimeTransaction.ApplyMutation(ct, createSprintParticipantMutation)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Sprint{}, err
		}
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Sprint{}, err
	}

	return sp, nil
}

func (s Sprint) DeleteSprint(ct context.Context, sprintID uint64) (entity.Sprint, error) {
	taskIds, err := s.sprintTaskRelationDao.FindTaskIDsBySprintID(ct, sprintID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Sprint{}, err
	}

	for _, taskId := range taskIds {
		_, err = s.RemoveTaskFromSprint(ct, sprintID, taskId)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Sprint{}, err
		}
	}

	participantUserIDs, err := s.sprintParticipantDao.FindParticipantIDsBySprintID(ct, sprintID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Sprint{}, err
	}

	sprint, err := s.sprintDao.FindSprintByID(ct, sprintID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Sprint{}, err
	}

	realTimeTransaction := realtime.NewTransaction(s.dataCollector, s.stateSyncer)
	for _, participantUserID := range participantUserIDs {
		deleteSprintParticipantMutation := mutation.NewDeleteSprintParticipantMutation(
			s.dataCollector,
			s.stateSyncer,
			s.sprintParticipantDao,
			s.sprintDao,
			participantUserID,
			sprintID)
		err = realTimeTransaction.ApplyMutation(ct, deleteSprintParticipantMutation)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Sprint{}, err
		}
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Sprint{}, err
	}

	return sprint, s.sprintDao.DeleteSprint(ct, sprintID)
}

func (s Sprint) AddTaskToSprint(ct context.Context, sprintID uint64, taskID uint64) (entity.Task, error) {
	task, err := s.taskDao.FindTaskByID(ct, taskID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	sprint, err := s.sprintDao.FindSprintByID(ct, sprintID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	if sprint.OwningTeamID != task.OwningTeamID {
		err = errors.New("sprint and task must belong to the same team")
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	realTimeTransaction := realtime.NewTransaction(s.dataCollector, s.stateSyncer)
	relation := entity.SprintTaskRelation{
		SprintID:  sprintID,
		TaskID:    taskID,
		CreatedAt: time.Now().UTC(),
	}
	createSprintTaskRelationMutation := mutation.NewCreateSprintTaskRelationMutation(
		s.dataCollector,
		s.stateSyncer,
		s.sprintTaskRelationDao,
		s.sprintDao,
		relation)
	err = realTimeTransaction.ApplyMutation(ct, createSprintTaskRelationMutation)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	if !task.IsPlanned {
		task.IsPlanned = true
		updateTaskMutation := mutation.NewUpdateTaskMutation(
			s.dataCollector,
			s.stateSyncer,
			s.taskDao,
			task)
		err = realTimeTransaction.ApplyMutation(ct, updateTaskMutation)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Task{}, err
		}
	}

	err = s.tryReduceBandwidth(ct, realTimeTransaction, sprintID, task)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	return task, nil
}

func (s Sprint) CopyTasksToSprint(ct context.Context, toSprintID uint64, taskIDs []uint64) ([]entity.Task, error) {
	if feature.EnableAuthorization {
		userID, ok := ctx.UserIDFromContext(ct)
		if !ok {
			err := errors.New("Unauthorized")
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return []entity.Task{}, err
		}

		sprint, err := s.sprintDao.FindSprintByID(ct, toSprintID)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return []entity.Task{}, err
		}

		query := authorization.NewCloneTaskQuery(userID, sprint.OwningTeamID)
		hasPermission, err := s.authorizer.hasPermission(ct, query)
		if err != nil {
			s.authorizer.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return []entity.Task{}, err
		}

		if !hasPermission {
			return []entity.Task{}, authorization.Error{
				Code:    authorization.UnauthorizedErrorCode,
				Message: fmt.Sprintf("Unauthorized: %v", query),
			}
		}
	}

	res := make([]entity.Task, 0)
	for _, taskID := range taskIDs {
		task, err := s.copyTaskToSprint(ct, toSprintID, taskID)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		res = append(res, task)
	}

	return res, nil
}

func (s Sprint) MoveTasksToSprint(ct context.Context, fromSprintID uint64, toSprintID uint64, taskIDs []uint64) ([]entity.Task, error) {
	res := make([]entity.Task, 0)
	for _, taskID := range taskIDs {
		task, err := s.moveTaskToSprint(ct, fromSprintID, toSprintID, taskID)

		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		res = append(res, task)
	}

	return res, nil
}

func (s Sprint) copyTaskToSprint(ct context.Context, toSprintID uint64, taskID uint64) (entity.Task, error) {
	task, err := s.taskDao.FindTaskByID(ct, taskID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	cloneTask := createTaskInput{
		Goal:          task.Goal,
		Context:       task.Context,
		OwnerUserID:   task.OwnerUserID,
		CreatorUserID: task.CreatorUserID,
		Status:        task.Status,
		DueAt:         task.DueAt,
		DeliveredAt:   task.DeliveredAt,
		IsPlanned:     task.IsPlanned,
		Effort:        task.Effort,
	}
	createdTask, err := s.taskService.createTask(ct, task.OwningTeamID, cloneTask)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	s.AddTaskToSprint(ct, toSprintID, createdTask.ID)

	return createdTask, nil
}

func (s Sprint) moveTaskToSprint(ct context.Context, fromSprintID uint64, toSprintID uint64, taskID uint64) (entity.Task, error) {
	sprintIDs, err := s.sprintTaskRelationDao.FindSprintIDsByTaskID(ct, taskID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	foundSprintIDs := collect.Filter(sprintIDs, func(currSprintID uint64) bool {
		return currSprintID == fromSprintID
	})
	if len(foundSprintIDs) < 1 {
		err = errors.New("relation not found")
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"SprintID": fromSprintID,
				"TaskID":   taskID,
			},
		})
		return entity.Task{}, err
	}

	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = s.sprintTaskRelationDao.DeleteSprintTaskRelation(ct, fromSprintID, taskID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	relation := entity.SprintTaskRelation{
		SprintID:  toSprintID,
		TaskID:    taskID,
		CreatedAt: time.Now().UTC(),
	}

	err = s.sprintTaskRelationDao.CreateSprintTaskRelation(ct, relation)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	task, err := s.taskDao.FindTaskByID(ct, taskID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	realTimeTransaction := realtime.NewTransaction(s.dataCollector, s.stateSyncer)
	err = s.tryIncreaseBandwidth(ct, realTimeTransaction, fromSprintID, task)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = s.tryReduceBandwidth(ct, realTimeTransaction, toSprintID, task)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	return task, nil
}

func (s Sprint) RemoveTaskFromSprint(ct context.Context, sprintID uint64, taskID uint64) (entity.Task, error) {
	sprintIDs, err := s.sprintTaskRelationDao.FindSprintIDsByTaskID(ct, taskID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	foundSprintIDs := collect.Filter(sprintIDs, func(currSprintID uint64) bool {
		return currSprintID == sprintID
	})
	if len(foundSprintIDs) < 1 {
		err = errors.New("relation not found")
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"SprintID": sprintID,
				"TaskID":   taskID,
			},
		})
		return entity.Task{}, err
	}

	task, err := s.taskDao.FindTaskByID(ct, taskID)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	realTimeTransaction := realtime.NewTransaction(s.dataCollector, s.stateSyncer)
	deleteSprintTaskRelationMutation := mutation.NewDeleteSprintTaskRelationMutation(
		s.dataCollector,
		s.stateSyncer,
		s.sprintTaskRelationDao,
		sprintID,
		task,
	)
	err = realTimeTransaction.ApplyMutation(ct, deleteSprintTaskRelationMutation)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}
	//if there is no other sprint that the task can move to,  put it into backlog
	if len(sprintIDs) <= 1 {
		task.IsPlanned = false
		updateTaskMutation := mutation.NewUpdateTaskMutation(
			s.dataCollector,
			s.stateSyncer,
			s.taskDao,
			task)
		err = realTimeTransaction.ApplyMutation(ct, updateTaskMutation)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Task{}, err
		}
	}

	err = s.tryIncreaseBandwidth(ct, realTimeTransaction, sprintID, task)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = realTimeTransaction.Commit(ct)
	if err != nil {
		s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	return task, nil
}

func (s Sprint) tryReduceBandwidth(ct context.Context, tx *realtime.Transaction, sprintID uint64, task entity.Task) error {
	if task.OwnerUserID != nil && task.Effort != nil {
		newSprintParticipant, err := s.sprintParticipantDao.FindParticipant(ct, sprintID, *task.OwnerUserID)
		if err != nil {
			// TODO: this should be removed once the sprint participants are backfilled
			if errors.As(err, &dao.ErrorNotFound) {
				return nil
			}

			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return err
		}

		newSprintParticipant.UnusedBandwidth -= *task.Effort
		updateSprintParticipantMutation := mutation.NewUpdateSprintParticipantMutation(
			s.dataCollector,
			s.stateSyncer,
			s.sprintParticipantDao,
			s.sprintDao,
			newSprintParticipant)
		err = tx.ApplyMutation(ct, updateSprintParticipantMutation)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return err
		}
	}

	return nil
}

func (s Sprint) tryIncreaseBandwidth(ct context.Context, tx *realtime.Transaction, sprintID uint64, task entity.Task) error {
	if task.OwnerUserID != nil && task.Effort != nil {
		oldSprintParticipant, err := s.sprintParticipantDao.FindParticipant(ct, sprintID, *task.OwnerUserID)
		if err != nil {
			// TODO: this should be removed once the sprint participants are backfilled
			if errors.As(err, &dao.ErrorNotFound) {
				return nil
			}

			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return err
		}

		oldSprintParticipant.UnusedBandwidth += *task.Effort
		updateSprintParticipantMutation := mutation.NewUpdateSprintParticipantMutation(
			s.dataCollector,
			s.stateSyncer,
			s.sprintParticipantDao,
			s.sprintDao,
			oldSprintParticipant)
		err = tx.ApplyMutation(ct, updateSprintParticipantMutation)
		if err != nil {
			s.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			return err
		}
	}

	return nil
}

func NewSprint(
	dataCollector obs.DataCollector,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	stateSyncer *realtime.StateSyncer,
	authorizer Authorizer,
	taskDao dao.Task,
	sprintDao dao.Sprint,
	sprintTaskRelationDao dao.SprintTaskRelation,
	sprintParticipantDao dao.SprintParticipant,
	teamMemberDao dao.TeamMember,
	taskService Task,
) Sprint {
	return Sprint{
		dataCollector:         dataCollector,
		cloudClientRegistry:   cloudClientRegistry,
		stateSyncer:           stateSyncer,
		authorizer:            authorizer,
		taskDao:               taskDao,
		sprintDao:             sprintDao,
		sprintTaskRelationDao: sprintTaskRelationDao,
		sprintParticipantDao:  sprintParticipantDao,
		teamMemberDao:         teamMemberDao,
		taskService:           taskService,
	}
}
