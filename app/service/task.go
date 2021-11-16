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

func (t Task) CreateTask(task entity.Task, userId oneEntity.ID) (oneEntity.ID, error) {
	//	1. create task
	//  2. add task to upcoming list of the current team
	id, err := t.taskRepo.CreateTask(task)
	if err != nil {
		log.Println(err)
		return 0, err
	}
	// find active team with this user
	// assign task to team
	activeTeam, err := t.teamRepo.GetActiveTeam(userId)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	err = t.taskRepo.AssignTaskToTeam(task.ID,activeTeam.ID, entity.TaskStatusUpcoming)
	if err != nil {
		log.Println(err)
		return 0, err
	}

	return id, nil
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

func NewTask(taskRepo repo.Task) Task {
	return Task{
		taskRepo: taskRepo,
	}
}
