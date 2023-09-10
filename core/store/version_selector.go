package store

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/rollout"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type ExperimentVersionSelector struct {
	rolloutViewerDao dao.RolloutViewer
	rolloutID        uint64
}

var _ rollout.ExperimentVersionSelectorStore = (*ExperimentVersionSelector)(nil)

func (e *ExperimentVersionSelector) GetViewerVersionNumber(ct context.Context, viewerID uint64) (*int, *errs.Error) {
	viewer, err := e.rolloutViewerDao.FindRolloutViewerByViewerIDAndRolloutID(ct, viewerID, e.rolloutID)
	return &viewer.VersionNumber, err
}

func (e *ExperimentVersionSelector) SetViewerVersionNumber(ct context.Context, viewerID uint64, versionNumber int) *errs.Error {
	viewer, err := e.rolloutViewerDao.FindRolloutViewerByViewerIDAndRolloutID(ct, viewerID, e.rolloutID)
	if err != nil {
		return err
	}

	viewer.VersionNumber = versionNumber
	return e.rolloutViewerDao.UpdateRolloutViewer(ct, viewer)
}

func NewExperimentVersionSelector(rolloutViewerDao dao.RolloutViewer, rolloutID uint64) *ExperimentVersionSelector {
	return &ExperimentVersionSelector{rolloutViewerDao: rolloutViewerDao, rolloutID: rolloutID}
}
