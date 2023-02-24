package daov2

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskAwaitForRelation interface {
	FindAwaitingTaskIDs(ct context.Context, tx *sql.Tx, waitForTaskID uint64) ([]uint64, *errs.Error)
	FindAwaitForTaskIDs(ct context.Context, tx *sql.Tx, waitingTaskID uint64) ([]uint64, *errs.Error)
	CreateRelation(ct context.Context, tx *sql.Tx, relation entity.TaskAwaitForRelation) *errs.Error
	DeleteRelation(ct context.Context, tx *sql.Tx, waitingTaskID uint64, awaitForTaskID uint64) *errs.Error
}
