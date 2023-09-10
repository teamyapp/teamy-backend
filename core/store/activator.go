package store

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/rollout"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type MaxViewersActivator struct {
	rolloutID        uint64
	rolloutViewerDao dao.RolloutViewer
	rolloutDao       dao.Rollout
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
	return m.rolloutViewerDao.UpdateRolloutViewer(ct, rolloutViewer)
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
	return m.rolloutDao.UpdateRollout(ct, rollout)
}

func NewMaxViewersActivator(rolloutViewerDao dao.RolloutViewer, rolloutDao dao.Rollout, rolloutID uint64) *MaxViewersActivator {
	return &MaxViewersActivator{rolloutID: rolloutID, rolloutViewerDao: rolloutViewerDao, rolloutDao: rolloutDao}
}

type PercentageActivator struct {
	rolloutViewerDao dao.RolloutViewer
	rolloutID        uint64
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
	return p.rolloutViewerDao.UpdateRolloutViewer(ct, viewer)
}

func NewPercentageActivator(rolloutViewerDao dao.RolloutViewer, rolloutID uint64) *PercentageActivator {
	return &PercentageActivator{rolloutViewerDao: rolloutViewerDao, rolloutID: rolloutID}
}
