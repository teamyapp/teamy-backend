package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Rollout struct{}

var _ dao.Rollout = (*Rollout)(nil)

func (*Rollout) FindRolloutByID(ct context.Context, rolloutIDs uint64) (entity.Rollout, *errs.Error) {
	panic("unimplemented")
}

func (*Rollout) FindRolloutByIDWithTx(ct context.Context, tx *transaction.Transaction, rolloutID uint64) (entity.Rollout, *errs.Error) {
	panic("unimplemented")
}

func (*Rollout) FindRolloutsByIDs(ct context.Context, rolloutIDs []uint64) ([]entity.Rollout, *errs.Error) {
	panic("unimplemented")
}

func (*Rollout) CreateRollout(ct context.Context, rollout entity.Rollout) (entity.Rollout, *errs.Error) {
	panic("unimplemented")
}

func (*Rollout) UpdateRollout(ct context.Context, rollout entity.Rollout) *errs.Error {
	panic("unimplemented")
}

func (*Rollout) UpdateRolloutWithTx(ct context.Context, tx *transaction.Transaction, rollout entity.Rollout) *errs.Error {
	panic("unimplemented")
}

func (*Rollout) DeleteRolloutWithTx(ct context.Context, tx *transaction.Transaction, rolloutID uint64) *errs.Error {
	panic("unimplemented")
}

func NewRollout() *Rollout {
	return &Rollout{}
}
