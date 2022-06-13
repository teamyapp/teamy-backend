package dao

import (
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Task interface {
	FindTaskByID(taskID uint64) (entity.Task, error)
	FindTasksByIDs(taskIDs []uint64) ([]entity.Task, error)
	FindTasksByTeamID(teamID uint64) ([]entity.Task, error)
	FindAllTasks() ([]entity.Task, error)
	CreateTask(task entity.Task) error
	UpdateTask(task entity.Task) error
	DeleteTask(taskID uint64) error
}
