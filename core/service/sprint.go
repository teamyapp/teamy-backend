package service

import (
	"context"
	"log"
	"time"

	cloudAPI "github.com/teamyapp/cloud/app/api"
	"github.com/teamyapp/cloud/app/api/proto"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const sprintLength = 7 * 24 * time.Hour

type CreateSprintInput struct {
	StartAt time.Time
}

type Sprint struct {
	cloudClientRegistry   *cloudAPI.ClientRegistry
	taskDao               dao.Task
	sprintDao             dao.Sprint
	sprintTaskRelationDao dao.SprintTaskRelation
}

func (s Sprint) FindTasksInSprint(
	ct context.Context,
	sprintID uint64,
	filter *TaskFilter,
) ([]entity.Task, error) {
	taskIDs, err := s.sprintTaskRelationDao.FindTaskIDsBySprintID(sprintID)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	tasks, err := s.taskDao.FindTasksByIDs(taskIDs)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if filter != nil {
		tasks = filterTasks(tasks, *filter)
	}

	return tasks, nil
}

func (s Sprint) FindSprints(ct context.Context, filter *SprintFilter) ([]entity.Sprint, error) {
	sprints, err := s.sprintDao.FindAllSprints()
	if err != nil {
		log.Println(err)
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
		log.Println(err)
		return entity.Sprint{}, err
	}

	sp := entity.Sprint{
		ID:           genSprintIDRes.UniqueNumber,
		StartAt:      sprint.StartAt.UTC(),
		EndAt:        sprint.StartAt.UTC().Add(sprintLength),
		CreatedAt:    time.Now().UTC(),
		OwningTeamID: teamID,
	}
	return sp, s.sprintDao.CreateSprint(sp)
}

func (s Sprint) DeleteSprint(ct context.Context, sprintID uint64) (entity.Sprint, error) {
	sprint, err := s.sprintDao.FindSprintByID(sprintID)
	if err != nil {
		log.Println(err)
		return entity.Sprint{}, err
	}

	return sprint, s.sprintDao.DeleteSprint(sprintID)
}

func (s Sprint) AddTaskToSprint(ct context.Context, sprintID uint64, taskID uint64) (entity.Task, error) {
	relation := entity.SprintTaskRelation{
		SprintID:  sprintID,
		TaskID:    taskID,
		CreatedAt: time.Now().UTC(),
	}

	err := s.sprintTaskRelationDao.CreateSprintTaskRelation(relation)
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}

	task, err := s.taskDao.FindTaskByID(taskID)
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}

	if task.IsPlanned {
		return task, nil
	}

	task.IsPlanned = true
	return task, s.taskDao.UpdateTask(task)
}

func (s Sprint) RemoveTaskFromSprint(ct context.Context, sprintID uint64, taskID uint64) (entity.Task, error) {
	err := s.sprintTaskRelationDao.DeleteSprintTaskRelation(sprintID, taskID)
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}

	task, err := s.taskDao.FindTaskByID(taskID)
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}

	sprintIDs, err := s.sprintTaskRelationDao.FindSprintIDsByTaskID(taskID)
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}

	if len(sprintIDs) > 0 {
		return task, nil
	}

	task.IsPlanned = false
	return task, s.taskDao.UpdateTask(task)
}

func NewSprint(
	cloudClientRegistry *cloudAPI.ClientRegistry,
	taskDao dao.Task,
	sprintDao dao.Sprint,
	sprintTaskRelationDao dao.SprintTaskRelation,
) Sprint {
	return Sprint{
		cloudClientRegistry:   cloudClientRegistry,
		taskDao:               taskDao,
		sprintDao:             sprintDao,
		sprintTaskRelationDao: sprintTaskRelationDao,
	}
}
