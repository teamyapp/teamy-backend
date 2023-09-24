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

type MaxViewersActivator struct {
	logger             telemetry.Logger
	transactionFactory cloudTransaction.Factory
	stateSyncer        *realtime.StateSyncer
	rolloutID          uint64
	rolloutViewerDao   dao.RolloutViewer
	rolloutDao         dao.Rollout
}

var _ rollout.MaxViewersActivatorStore = (*MaxViewersActivator)(nil)

func (m *MaxViewersActivator) GetIsActivated(ct context.Context, viewerID uint64) (*bool, *errs.Error) {
	rolloutViewer, err := m.rolloutViewerDao.FindRolloutViewerByViewerIDAndRolloutID(ct, viewerID, m.rolloutID)
	if err != nil {
		if err.Code == errs.NotFound {
			return nil, nil
		}

		return nil, err
	}

	return &rolloutViewer.IsActivated, err
}

func (m *MaxViewersActivator) SetIsActivated(ct context.Context, viewerID uint64, isActivated bool) *errs.Error {
	rolloutViewer, err := m.rolloutViewerDao.FindRolloutViewerByViewerIDAndRolloutID(ct, viewerID, m.rolloutID)
	if err != nil {
		return err
	}

	rolloutViewer.IsActivated = isActivated
	txCtx := transaction.NewTransactionsContext(
		m.logger,
		m.transactionFactory,
		m.stateSyncer,
		ct,
	)
	return txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return m.rolloutViewerDao.UpdateRolloutViewer(ct, tx, rolloutViewer)
	})
}

func (m *MaxViewersActivator) GetTotalViewers(ct context.Context, defaultViewers int) (int, *errs.Error) {
	rollout, err := m.rolloutDao.FindRolloutByID(ct, m.rolloutID)
	return rollout.Viewers, err
}

func (m *MaxViewersActivator) SetTotalViewers(ct context.Context, totalViewers int) *errs.Error {
	rollout, err := m.rolloutDao.FindRolloutByID(ct, m.rolloutID)
	if err != nil {
		return err
	}

	rollout.Viewers = totalViewers
	txCtx := transaction.NewTransactionsContext(
		m.logger,
		m.transactionFactory,
		m.stateSyncer,
		ct,
	)
	return txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return m.rolloutDao.UpdateRollout(ct, tx, rollout)
	})
}

func NewMaxViewersActivator(
	logger telemetry.Logger,
	transactionFactory cloudTransaction.Factory,
	stateSyncer *realtime.StateSyncer,
	rolloutViewerDao dao.RolloutViewer,
	rolloutDao dao.Rollout,
	rolloutID uint64,
) *MaxViewersActivator {
	return &MaxViewersActivator{
		logger:             logger,
		transactionFactory: transactionFactory,
		stateSyncer:        stateSyncer,
		rolloutID:          rolloutID,
		rolloutViewerDao:   rolloutViewerDao,
		rolloutDao:         rolloutDao,
	}
}

type PercentageActivator struct {
	logger             telemetry.Logger
	transactionFactory cloudTransaction.Factory
	stateSyncer        *realtime.StateSyncer
	rolloutViewerDao   dao.RolloutViewer
	rolloutID          uint64
}

var _ rollout.PercentageActivatorStore = (*PercentageActivator)(nil)

func (p *PercentageActivator) GetIsActivated(ct context.Context, viewerID uint64) (*bool, *errs.Error) {
	viewer, err := p.rolloutViewerDao.FindRolloutViewerByViewerIDAndRolloutID(ct, viewerID, p.rolloutID)
	if err != nil {
		if err.Code == errs.NotFound {
			return nil, nil
		}

		return nil, err
	}

	return &viewer.IsActivated, nil
}

func (p *PercentageActivator) SetIsActivated(ct context.Context, viewerID uint64, isActivated bool) *errs.Error {
	viewer, err := p.rolloutViewerDao.FindRolloutViewerByViewerIDAndRolloutID(ct, viewerID, p.rolloutID)
	if err != nil {
		return err
	}

	viewer.IsActivated = isActivated
	txCtx := transaction.NewTransactionsContext(
		p.logger,
		p.transactionFactory,
		p.stateSyncer,
		ct,
	)
	return txCtx.WithTransactions(false, func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
		return p.rolloutViewerDao.UpdateRolloutViewer(ct, tx, viewer)
	})
}

func NewPercentageActivator(
	logger telemetry.Logger,
	transactionFactory cloudTransaction.Factory,
	stateSyncer *realtime.StateSyncer,
	rolloutViewerDao dao.RolloutViewer,
	rolloutID uint64,
) *PercentageActivator {
	return &PercentageActivator{
		logger:             logger,
		transactionFactory: transactionFactory,
		stateSyncer:        stateSyncer,
		rolloutViewerDao:   rolloutViewerDao,
		rolloutID:          rolloutID,
	}
}
