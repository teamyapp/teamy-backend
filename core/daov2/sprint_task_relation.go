package daov2

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintTaskRelation interface {
	FindTaskIDsBySprintID(ct context.Context, tx *transaction.Transaction, sprintID uint64) ([]uint64, *errs.Error)
	FindSprintIDsByTaskID(ct context.Context, tx *transaction.Transaction, taskID uint64) ([]uint64, *errs.Error)
	CreateSprintTaskRelation(ct context.Context, tx *transaction.Transaction, relation entity.SprintTaskRelation) *errs.Error
	DeleteSprintTaskRelation(ct context.Context, tx *transaction.Transaction, sprintID uint64, taskID uint64) *errs.Error
}
