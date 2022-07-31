package dao

import (
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Sprint interface {
	FindSprintByID(sprintID uint64) (entity.Sprint, error)
	FindSprintsByIDs(sprintIDs []uint64) ([]entity.Sprint, error)
	FindSprintsByTeamID(teamID uint64) ([]entity.Sprint, error)
	FindAllSprints() ([]entity.Sprint, error)
	CreateSprint(sprint entity.Sprint) error
	DeleteSprint(sprintID uint64) error
}
