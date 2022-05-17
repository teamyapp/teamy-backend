package dao

import (
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type Task interface {
	FindTasksByTeamID(teamID uint64) ([]entityv2.Task, error)
	FindAllTasks() ([]entityv2.Task, error)
}
