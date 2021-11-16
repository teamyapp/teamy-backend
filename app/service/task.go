package service

import (
	"log"

	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
	"github.com/teamyapp/teamy-backend/app/repo"
)

type Task struct {
	taskRepo repo.Task
	teamRepo repo.Team
}

func (t Task) FindTask(taskID oneEntity.ID) (entity.Task, error) {
	task, err := t.taskRepo.FindTaskByID(taskID)
	if err != nil {
		log.Println(err)
		return entity.Task{}, err
	}
	return task, nil
}

func (t Task) CreateTask(task entity.Task, userId oneEntity.ID) error {
	_, err := t.taskRepo.CreateTask(task)
	if err != nil {
		log.Println(err)
	}

	activeTeam, err := t.teamRepo.GetActiveTeam(userId)
	if err != nil {
		log.Println(err)
		return err
	}

	err = t.taskRepo.AssignTaskToTeam(task.ID, activeTeam.ID, entity.TaskStatusUpcoming)
	if err != nil {
		log.Println(err)
		return err
	}

	return nil
}

func (t Task) DeleteTask(taskID oneEntity.ID) error {
	panic("not implemented")
}

func (t Task) UpdateTask(task entity.Task) error {
	panic("not implemented")
}

func (t Task) PerformTaskAction(taskID oneEntity.ID, action entity.TaskAction) error {
	panic("not implemented")
}

func NewTask(taskRepo repo.Task, teamRepo repo.Team) Task {
	return Task{
		taskRepo: taskRepo,
		teamRepo: teamRepo,
	}
}
