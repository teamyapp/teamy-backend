package service

import (
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Task struct {
}

func (t Task) GetTask(taskID int) (entity.Task, error) {
	panic("not implemented")
}

func (t Task) CreateTask(task entity.Task) error {
	panic("not implemented")
}

func (t Task) DeleteTask(taskID int) error {
	panic("not implemented")
}

func (t Task) UpdateTask(task entity.Task) error {
	panic("not implemented")
}

func (t Task) PerformTaskAction(taskID int, action entity.TaskAction) error {
	panic("not implemented")
}
