package dao

import (
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type Task interface {
	FindTaskByID(taskID uint64) (entityv2.Task, error)
	FindTasksByTeamID(teamID uint64) ([]entityv2.Task, error)
	FindAllTasks() ([]entityv2.Task, error)
	CreateTask(task entityv2.Task) error
	UpdateTask(task entityv2.Task) error
}
