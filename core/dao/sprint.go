package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/core/entity"
)

type Sprint interface {
	FindSprintByID(ct context.Context, sprintID uint64) (entity.Sprint, error)
	FindSprintsByIDs(ct context.Context, sprintIDs []uint64) ([]entity.Sprint, error)
	FindSprintsByTeamID(ct context.Context, teamID uint64) ([]entity.Sprint, error)
	FindAllSprints(ct context.Context) ([]entity.Sprint, error)
	CreateSprint(ct context.Context, sprint entity.Sprint) error
	DeleteSprint(ct context.Context, sprintID uint64) error
}
