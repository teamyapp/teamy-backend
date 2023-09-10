package service

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/rollout"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type MaxViewersActivatorStore struct {
	rolloutID        uint64
	rolloutViewerDao dao.RolloutViewer
	rolloutStoreDao  dao.RolloutStore
}

var _ rollout.MaxViewersActivatorStore = (*MaxViewersActivatorStore)(nil)

func (m *MaxViewersActivatorStore) GetIsActivated(ct context.Context, viewerID uint64) (*bool, *errs.Error) {
	rolloutViewer, err := m.rolloutViewerDao.FindRolloutViewerByViewerIDAndRolloutID(ct, viewerID, m.rolloutID)
	if err != nil {
		if err.Code == errs.NotFound {
			return nil, nil
		}

		return nil, err
	}

	return &rolloutViewer.IsActivated, err
}

func (m *MaxViewersActivatorStore) SetIsActivated(ct context.Context, viewerID uint64, isActivated bool) *errs.Error {
	rolloutViewer, err := m.rolloutViewerDao.FindRolloutViewerByViewerIDAndRolloutID(ct, viewerID, m.rolloutID)
	if err != nil {
		return err
	}

	rolloutViewer.IsActivated = isActivated
	return m.rolloutViewerDao.UpdateRolloutViewer(ct, rolloutViewer)
}

func (m *MaxViewersActivatorStore) GetTotalViewers(ct context.Context, defaultViewers int) (int, *errs.Error) {
	store, err := m.rolloutStoreDao.FindRolloutStoreByID(ct, m.rolloutID)
	return store.TotalViewers, err
}

func (m *MaxViewersActivatorStore) SetTotalViewers(ct context.Context, totalViewers int) *errs.Error {
	store, err := m.rolloutStoreDao.FindRolloutStoreByID(ct, m.rolloutID)
	if err != nil {
		return err
	}

	store.TotalViewers = totalViewers
	return m.rolloutStoreDao.UpdateRolloutStore(ct, store)
}

func NewMaxViewersActivatorStore(rolloutViewerDao dao.RolloutViewer, rolloutStoreDao dao.RolloutStore, rolloutID uint64) *MaxViewersActivatorStore {
	return &MaxViewersActivatorStore{rolloutID: rolloutID, rolloutViewerDao: rolloutViewerDao, rolloutStoreDao: rolloutStoreDao}
}

type ExperimentVersionSelectorStore struct {
	rolloutViewerDao dao.RolloutViewer
	rolloutID        uint64
}

var _ rollout.ExperimentVersionSelectorStore = (*ExperimentVersionSelectorStore)(nil)

func (e *ExperimentVersionSelectorStore) GetViewerVersionNumber(ct context.Context, viewerID uint64) (*int, *errs.Error) {
	viewer, err := e.rolloutViewerDao.FindRolloutViewerByViewerIDAndRolloutID(ct, viewerID, e.rolloutID)
	return &viewer.VersionNumber, err
}

func (e *ExperimentVersionSelectorStore) SetViewerVersionNumber(ct context.Context, viewerID uint64, versionNumber int) *errs.Error {
	viewer, err := e.rolloutViewerDao.FindRolloutViewerByViewerIDAndRolloutID(ct, viewerID, e.rolloutID)
	if err != nil {
		return err
	}

	viewer.VersionNumber = versionNumber
	return e.rolloutViewerDao.UpdateRolloutViewer(ct, viewer)
}

func NewExperimentVersionSelectorStore(rolloutViewerDao dao.RolloutViewer, rolloutID uint64) *ExperimentVersionSelectorStore {
	return &ExperimentVersionSelectorStore{rolloutViewerDao: rolloutViewerDao, rolloutID: rolloutID}
}

type PercentageActivatorStore struct {
	rolloutViewerDao dao.RolloutViewer
	rolloutID        uint64
}

var _ rollout.PercentageActivatorStore = (*PercentageActivatorStore)(nil)

func (p *PercentageActivatorStore) GetIsActivated(ct context.Context, viewerID uint64) (*bool, *errs.Error) {
	viewer, err := p.rolloutViewerDao.FindRolloutViewerByViewerIDAndRolloutID(ct, viewerID, p.rolloutID)
	if err != nil {
		if err.Code == errs.NotFound {
			return nil, nil
		}

		return nil, err
	}

	return &viewer.IsActivated, nil
}

func (p *PercentageActivatorStore) SetIsActivated(ct context.Context, viewerID uint64, isActivated bool) *errs.Error {
	viewer, err := p.rolloutViewerDao.FindRolloutViewerByViewerIDAndRolloutID(ct, viewerID, p.rolloutID)
	if err != nil {
		return err
	}

	viewer.IsActivated = isActivated
	return p.rolloutViewerDao.UpdateRolloutViewer(ct, viewer)
}

func NewPercentageActivatorStore(rolloutViewerDao dao.RolloutViewer, rolloutID uint64) *PercentageActivatorStore {
	return &PercentageActivatorStore{rolloutViewerDao: rolloutViewerDao, rolloutID: rolloutID}
}

type RolloutStore struct {
	rolloutDao dao.Rollout
	rolloutID  uint64
}

var _ rollout.Store = (*RolloutStore)(nil)

func (r *RolloutStore) GetIsRolloutEnabled(ct context.Context, defaultIsRolloutEnabled bool) (bool, *errs.Error) {
	rawRollout, err := r.rolloutDao.FindRolloutByID(ct, r.rolloutID)
	return rawRollout.IsEnabled, err

}

func (r *RolloutStore) SetIsRolloutEnabled(ct context.Context, isRolloutEnabled bool) *errs.Error {
	rawRollout, err := r.rolloutDao.FindRolloutByID(ct, r.rolloutID)
	if err != nil {
		return err
	}

	rawRollout.IsEnabled = isRolloutEnabled
	return r.rolloutDao.UpdateRollout(ct, rawRollout)
}

func NewRolloutStore(
	rolloutDao dao.Rollout,
	rolloutID uint64,
) *RolloutStore {
	return &RolloutStore{rolloutDao: rolloutDao, rolloutID: rolloutID}
}
