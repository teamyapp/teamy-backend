package service

import (
	"context"
	"errors"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/cloud/libs/collect"
	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/collection"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type CreateSprintInput struct {
	StartAt time.Time
	EndAt   time.Time
}

type Sprint struct {
	dataCollector            obs.DataCollector
	cloudClientRegistry      *cloudAPI.ClientRegistry
	taskDao                  dao.Task
	sprintDao                dao.Sprint
	sprintTaskRelationDao    dao.SprintTaskRelation
	taskSyncer               collection.TaskSyncer
	sprintTaskRelationSyncer collection.SprintTaskRelationSyncer
}

func (s Sprint) FindTasksInSprint(
	ct context.Context,
	sprintID uint64,
	filter *TaskFilter,
) ([]entity.Task, error) {
	taskIDs, err := s.sprintTaskRelationDao.FindTaskIDsBySprintID(sprintID)
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	tasks, err := s.taskDao.FindTasksByIDs(taskIDs)
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	if filter != nil {
		tasks = filterTasks(tasks, *filter)
	}

	return tasks, nil
}

func (s Sprint) FindCurrentSprint(ct context.Context, teamID uint64) (entity.Sprint, error) {
	sprints, err := s.sprintDao.FindSprintsByTeamID(teamID)
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Sprint{}, err
	}

	now := time.Now().UTC()
	sprints = collect.Filter(sprints, func(sprint entity.Sprint) bool {
		if now.Before(sprint.StartAt) || now.After(sprint.EndAt) {
			return false
		}

		return true
	})
	if len(sprints) < 1 {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"teamID":      teamID,
				"currentTime": now.UTC(),
			},
		})
		return entity.Sprint{}, err
	}

	if len(sprints) > 1 {
		err = errors.New("team has more than one sprint")
		s.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"teamID": teamID,
			},
		})
		return entity.Sprint{}, err
	}

	return sprints[0], nil
}

func (s Sprint) FindSprints(ct context.Context, filter *SprintFilter) ([]entity.Sprint, error) {
	sprints, err := s.sprintDao.FindAllSprints()
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
	}

	if filter != nil {
		sprints = filterSprints(sprints, *filter)
	}

	return sprints, nil
}

func (s Sprint) CreateSprint(ct context.Context, teamID uint64, sprint CreateSprintInput) (entity.Sprint, error) {
	genSprintIDReq := &proto.GenerateUniqueNumberRequest{SequenceName: "sprintID"}
	genSprintIDRes, err := s.cloudClientRegistry.GeneratorClient().GenerateUniqueNumber(ct, genSprintIDReq)
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Sprint{}, err
	}

	sp := entity.Sprint{
		ID:           genSprintIDRes.UniqueNumber,
		StartAt:      sprint.StartAt.UTC(),
		EndAt:        sprint.EndAt.UTC(),
		CreatedAt:    time.Now().UTC(),
		OwningTeamID: teamID,
	}
	return sp, s.sprintDao.CreateSprint(sp)
}

func (s Sprint) DeleteSprint(ct context.Context, sprintID uint64) (entity.Sprint, error) {
	taskIds, err := s.sprintTaskRelationDao.FindTaskIDsBySprintID(sprintID)
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Sprint{}, err
	}

	for _, taskId := range taskIds {
		_, err = s.RemoveTaskFromSprint(ct, sprintID, taskId)
		if err != nil {
			s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			return entity.Sprint{}, err
		}
	}

	sprint, err := s.sprintDao.FindSprintByID(sprintID)
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Sprint{}, err
	}

	return sprint, s.sprintDao.DeleteSprint(sprintID)
}

func (s Sprint) AddTaskToSprint(ct context.Context, sprintID uint64, taskID uint64) (entity.Task, error) {
	task, err := s.taskDao.FindTaskByID(taskID)
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	sprint, err := s.sprintDao.FindSprintByID(sprintID)
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	if sprint.OwningTeamID != task.OwningTeamID {
		err = errors.New("sprint and task must belong to the same team")
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	relation := entity.SprintTaskRelation{
		SprintID:  sprintID,
		TaskID:    taskID,
		CreatedAt: time.Now().UTC(),
	}
	err = s.sprintTaskRelationSyncer.CreateAndSyncSprintTaskRelation(relation, task.OwningTeamID)
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	if task.IsPlanned {
		return task, nil
	}

	task.IsPlanned = true
	return task, s.taskSyncer.UpdateAndSyncTask(task)
}

func (s Sprint) MoveTasksToSprint(ct context.Context, fromSprintID uint64, toSprintID uint64, taskIDs []uint64) ([]entity.Task, error) {
	res := make([]entity.Task, 0)
	for _, taskID := range taskIDs {
		task, err := s.moveTaskToSprint(ct, fromSprintID, toSprintID, taskID)

		if err != nil {
			s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		res = append(res, task)
	}

	return res, nil
}

func (s Sprint) moveTaskToSprint(ct context.Context, fromSprintID uint64, toSprintID uint64, taskID uint64) (entity.Task, error) {
	sprintIDs, err := s.sprintTaskRelationDao.FindSprintIDsByTaskID(taskID)
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	foundSprintIDs := collect.Filter(sprintIDs, func(currSprintID uint64) bool {
		return currSprintID == fromSprintID
	})
	if len(foundSprintIDs) < 1 {
		err = errors.New("relation not found")
		s.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"sprintID": fromSprintID,
				"taskID":   taskID,
			},
		})
		return entity.Task{}, err
	}

	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = s.sprintTaskRelationDao.DeleteSprintTaskRelation(fromSprintID, taskID)
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	relation := entity.SprintTaskRelation{
		SprintID:  toSprintID,
		TaskID:    taskID,
		CreatedAt: time.Now().UTC(),
	}

	err = s.sprintTaskRelationDao.CreateSprintTaskRelation(relation)
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	task, err := s.taskDao.FindTaskByID(taskID)
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	return task, nil
}

func (s Sprint) RemoveTaskFromSprint(ct context.Context, sprintID uint64, taskID uint64) (entity.Task, error) {
	sprintIDs, err := s.sprintTaskRelationDao.FindSprintIDsByTaskID(taskID)
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	foundSprintIDs := collect.Filter(sprintIDs, func(currSprintID uint64) bool {
		return currSprintID == sprintID
	})
	if len(foundSprintIDs) < 1 {
		err = errors.New("relation not found")
		s.dataCollector.Logger.Log(obs.Error, obs.Props{
			obs.CauseProp: err,
			obs.MessageProp: obs.Props{
				"sprintID": sprintID,
				"taskID":   taskID,
			},
		})
		return entity.Task{}, err
	}

	task, err := s.taskDao.FindTaskByID(taskID)
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	err = s.sprintTaskRelationSyncer.DeleteAndSyncSprintTaskRelation(sprintID, taskID, task.OwningTeamID)
	if err != nil {
		s.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
		return entity.Task{}, err
	}

	if len(sprintIDs) > 1 {
		return task, nil
	}

	task.IsPlanned = false
	return task, s.taskSyncer.UpdateAndSyncTask(task)
}

func NewSprint(
	dataCollector obs.DataCollector,
	cloudClientRegistry *cloudAPI.ClientRegistry,
	taskDao dao.Task,
	sprintDao dao.Sprint,
	sprintTaskRelationDao dao.SprintTaskRelation,
	taskSyncer collection.TaskSyncer,
	sprintTaskRelationSyncer collection.SprintTaskRelationSyncer,
) Sprint {
	return Sprint{
		dataCollector:            dataCollector,
		cloudClientRegistry:      cloudClientRegistry,
		taskDao:                  taskDao,
		sprintDao:                sprintDao,
		sprintTaskRelationDao:    sprintTaskRelationDao,
		taskSyncer:               taskSyncer,
		sprintTaskRelationSyncer: sprintTaskRelationSyncer,
	}
}
