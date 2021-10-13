package service

import (
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Task struct {
}

func (t Task) GetTask(taskID entity.ID) (entity.Task, error) {
	panic("not implemented")
}

func (t Task) CreateTask(task entity.Task) error {
	panic("not implemented")
}

func (t Task) DeleteTask(taskID entity.ID) error {
	panic("not implemented")
}

func (t Task) UpdateTask(task entity.Task) error {
	panic("not implemented")
}

func (t Task) PerformTaskAction(taskID entity.ID, action entity.TaskAction) error {
	panic("not implemented")
}
