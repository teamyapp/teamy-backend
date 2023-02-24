package daov2

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintTaskRelation interface {
	FindTaskIDsBySprintID(ct context.Context, tx *sql.Tx, sprintID uint64) ([]uint64, *errs.Error)
	FindSprintIDsByTaskID(ct context.Context, tx *sql.Tx, taskID uint64) ([]uint64, *errs.Error)
	CreateSprintTaskRelation(ct context.Context, tx *sql.Tx, relation entity.SprintTaskRelation) *errs.Error
	DeleteSprintTaskRelation(ct context.Context, tx *sql.Tx, sprintID uint64, taskID uint64) *errs.Error
}
