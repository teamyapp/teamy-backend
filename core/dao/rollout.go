package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Rollout interface {
	FindRolloutByID(ct context.Context, rolloutIDs uint64) (entity.Rollout, *errs.Error)
	FindRolloutByIDWithTx(ct context.Context, tx *transaction.Transaction, rolloutID uint64) (entity.Rollout, *errs.Error)
	FindRolloutsByIDs(ct context.Context, rolloutIDs []uint64) ([]entity.Rollout, *errs.Error)
	CreateRollout(ct context.Context, tx *transaction.Transaction, rollout entity.Rollout) *errs.Error
	UpdateRollout(ct context.Context, tx *transaction.Transaction, rollout entity.Rollout) *errs.Error
	DeleteRollout(ct context.Context, tx *transaction.Transaction, rolloutID uint64) *errs.Error
}
