package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Rollout struct{}

// UpdateRolloutWithTx implements dao.Rollout.
func (*Rollout) UpdateRolloutWithTx(ct context.Context, tx *transaction.Transaction, rollout entity.Rollout) *errs.Error {
	panic("unimplemented")
}

// CreateRollout implements dao.Rollout.
func (*Rollout) CreateRollout(ct context.Context, rollout entity.Rollout) (entity.Rollout, *errs.Error) {
	panic("unimplemented")
}

// DeleteRolloutWithTx implements dao.Rollout.
func (*Rollout) DeleteRolloutWithTx(ct context.Context, tx *transaction.Transaction, rolloutID uint64) *errs.Error {
	panic("unimplemented")
}

// FindRolloutByID implements dao.Rollout.
func (*Rollout) FindRolloutByID(ct context.Context, rolloutIDs uint64) (entity.Rollout, *errs.Error) {
	panic("unimplemented")
}

// FindRolloutByIDWithTx implements dao.Rollout.
func (*Rollout) FindRolloutByIDWithTx(ct context.Context, tx *transaction.Transaction, rolloutID uint64) (entity.Rollout, *errs.Error) {
	panic("unimplemented")
}

// FindRolloutsByIDs implements dao.Rollout.
func (*Rollout) FindRolloutsByIDs(ct context.Context, rolloutIDs []uint64) ([]entity.Rollout, *errs.Error) {
	panic("unimplemented")
}

// UpdateRollout implements dao.Rollout.
func (*Rollout) UpdateRollout(ct context.Context, rollout entity.Rollout) *errs.Error {
	panic("unimplemented")
}

var _ dao.Rollout = (*Rollout)(nil)

func NewRollout() *Rollout {
	return &Rollout{}
}
