package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Sprint interface {
	FindSprintByID(ct context.Context, sprintID uint64) (entity.Sprint, *errs.Error)
	FindSprintsByIDs(ct context.Context, sprintIDs []uint64) ([]entity.Sprint, *errs.Error)
	FindSprintsByTeamID(ct context.Context, teamID uint64) ([]entity.Sprint, *errs.Error)
	FindAllSprints(ct context.Context) ([]entity.Sprint, *errs.Error)
	CreateSprint(ct context.Context, sprint entity.Sprint) *errs.Error
	DeleteSprint(ct context.Context, sprintID uint64) *errs.Error
}
