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
	logger             telemetry.Logger
	transactionFactory cloudTransaction.Factory
	stateSyncer        *realtime.StateSyncer
	rolloutDao         dao.Rollout
	rolloutID          uint64
}

var _ rollout.Store = (*Rollout)(nil)

func (r *Rollout) GetIsRolloutEnabled(ct context.Context, defaultIsRolloutEnabled bool) (bool, *errs.Error) {
	rollout, err := r.rolloutDao.FindRolloutByID(ct, r.rolloutID)
	return rollout.IsEnabled, err

}

func (r *Rollout) SetIsRolloutEnabled(ct context.Context, isRolloutEnabled bool) *errs.Error {
	txCtx := transaction.NewTransactionsContext(
		r.logger,
		r.transactionFactory,
		r.stateSyncer,
		ct,
	)
	return txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		rollout, err := r.rolloutDao.FindRolloutByID(ct, r.rolloutID)
		if err != nil {
			return err
		}

		rollout.IsEnabled = isRolloutEnabled
		return r.rolloutDao.UpdateRollout(ct, tx, rollout)
	})
}

func NewRollout(
	logger telemetry.Logger,
	transactionFactory cloudTransaction.Factory,
	stateSyncer *realtime.StateSyncer,
	rolloutDao dao.Rollout,
	rolloutID uint64,
) *Rollout {
	return &Rollout{
		logger:             logger,
		transactionFactory: transactionFactory,
		stateSyncer:        stateSyncer,
		rolloutDao:         rolloutDao,
		rolloutID:          rolloutID,
	}
}
