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

type ExperimentVersionSelector struct {
	logger                  telemetry.Logger
	transactionGroupFactory transaction.GroupFactory
	transactionFactory      cloudTransaction.Factory
	stateSyncer             *realtime.StateSyncer
	rolloutViewerDao        dao.RolloutViewer
	rolloutID               uint64
}

var _ rollout.ExperimentVersionSelectorStore = (*ExperimentVersionSelector)(nil)

func (e *ExperimentVersionSelector) GetViewerVersionNumber(ct context.Context, viewerID uint64) (*int, *errs.Error) {
	viewer, err := e.rolloutViewerDao.FindRolloutViewerByViewerIDAndRolloutID(ct, viewerID, e.rolloutID)
	return &viewer.VersionNumber, err
}

func (e *ExperimentVersionSelector) SetViewerVersionNumber(ct context.Context, viewerID uint64, versionNumber int) *errs.Error {
	return e.transactionGroupFactory.WithTransactionGroup(
		ct,
		false,
		func(tx *cloudTransaction.Transaction, rtTx *realtime.Transaction) *errs.Error {
			viewer, err := e.rolloutViewerDao.FindRolloutViewerByViewerIDAndRolloutID(ct, viewerID, e.rolloutID)
			if err != nil {
				return err
			}

			viewer.VersionNumber = versionNumber
			return e.rolloutViewerDao.UpdateRolloutViewer(ct, tx, viewer)
		})
}

func NewExperimentVersionSelector(
	logger telemetry.Logger,
	transactionGroupFactory transaction.GroupFactory,
	transactionFactory cloudTransaction.Factory,
	stateSyncer *realtime.StateSyncer,
	rolloutViewerDao dao.RolloutViewer,
	rolloutID uint64,
) *ExperimentVersionSelector {
	return &ExperimentVersionSelector{
		logger:                  logger,
		transactionGroupFactory: transactionGroupFactory,
		transactionFactory:      transactionFactory,
		stateSyncer:             stateSyncer,
		rolloutViewerDao:        rolloutViewerDao,
		rolloutID:               rolloutID,
	}
}
