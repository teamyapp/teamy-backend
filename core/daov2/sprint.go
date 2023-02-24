package daov2

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Sprint interface {
	FindSprintByID(ct context.Context, tx *sql.Tx, sprintID uint64) (entity.Sprint, *errs.Error)
	FindSprintsByIDs(ct context.Context, tx *sql.Tx, sprintIDs []uint64) ([]entity.Sprint, *errs.Error)
	FindSprintsByTeamID(ct context.Context, tx *sql.Tx, teamID uint64) ([]entity.Sprint, *errs.Error)
	FindAllSprints(ct context.Context, tx *sql.Tx) ([]entity.Sprint, *errs.Error)
	CreateSprint(ct context.Context, tx *sql.Tx, sprint entity.Sprint) *errs.Error
	DeleteSprint(ct context.Context, tx *sql.Tx, sprintID uint64) *errs.Error
}
