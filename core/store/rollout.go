package store

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/rollout"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type Rollout struct {
	rolloutDao dao.Rollout
	rolloutID  uint64
}

var _ rollout.Store = (*Rollout)(nil)

func (r *Rollout) GetIsRolloutEnabled(ct context.Context, defaultIsRolloutEnabled bool) (bool, *errs.Error) {
	rollout, err := r.rolloutDao.FindRolloutByID(ct, r.rolloutID)
	return rollout.IsEnabled, err

}

func (r *Rollout) SetIsRolloutEnabled(ct context.Context, isRolloutEnabled bool) *errs.Error {
	rollout, err := r.rolloutDao.FindRolloutByID(ct, r.rolloutID)
	if err != nil {
		return err
	}

	rollout.IsEnabled = isRolloutEnabled
	return r.rolloutDao.UpdateRollout(ct, rollout)
}

func NewRollout(
	rolloutDao dao.Rollout,
	rolloutID uint64,
) *Rollout {
	return &Rollout{rolloutDao: rolloutDao, rolloutID: rolloutID}
}
