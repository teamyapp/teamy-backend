package store

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/rollout"
	"github.com/teamyapp/cloud/libs/telemetry"
	cloudTransaction "github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/realtime"
	"github.com/teamyapp/teamy-backend/core/transaction"
)

type Rollout struct {
	logger                  telemetry.Logger
	transactionGroupFactory transaction.GroupFactory
	transactionFactory      cloudTransaction.Factory
	stateSyncer             *realtime.StateSyncer
	rolloutDao              dao.Rollout
	rolloutID               uint64
}

var _ rollout.Store = (*Rollout)(nil)

func (r *Rollout) GetIsRolloutEnabled(ct context.Context, defaultIsRolloutEnabled bool) (bool, *errs.Error) {
	ro, err := r.rolloutDao.FindRolloutByID(ct, r.rolloutID)
	return ro.IsEnabled, err

}

func (r *Rollout) SetIsRolloutEnabled(ct context.Context, isRolloutEnabled bool) *errs.Error {
	return r.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			ro, err := r.rolloutDao.FindRolloutByID(ct, r.rolloutID)
			if err != nil {
				return err
			}

			ro.IsEnabled = isRolloutEnabled
			return r.rolloutDao.UpdateRollout(ct, tx, ro)
		})
}

func NewRollout(
	logger telemetry.Logger,
	transactionGroupFactory transaction.GroupFactory,
	transactionFactory cloudTransaction.Factory,
	stateSyncer *realtime.StateSyncer,
	rolloutDao dao.Rollout,
	rolloutID uint64,
) *Rollout {
	return &Rollout{
		logger:                  logger,
		transactionGroupFactory: transactionGroupFactory,
		transactionFactory:      transactionFactory,
		stateSyncer:             stateSyncer,
		rolloutDao:              rolloutDao,
		rolloutID:               rolloutID,
	}
}
