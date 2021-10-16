package service

import (
	oneEntity "github.com/teamyapp/one/entity"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Task struct {
}

func (t Task) GetTask(taskID oneEntity.ID) (entity.Task, error) {
	panic("not implemented")
}

func (t Task) CreateTask(task entity.Task) error {
	panic("not implemented")
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
