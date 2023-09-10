package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type RolloutViewer struct{}

// CreateRolloutViewer implements dao.RolloutViewer.
func (*RolloutViewer) CreateRolloutViewer(ct context.Context, viewer entity.RolloutViewer) (entity.RolloutViewer, *errs.Error) {
	panic("unimplemented")
}

// FindRolloutViewerByViewerIDAndRolloutID implements dao.RolloutViewer.
func (*RolloutViewer) FindRolloutViewerByViewerIDAndRolloutID(ct context.Context, viewerID uint64, RolloutID uint64) (entity.RolloutViewer, *errs.Error) {
	panic("unimplemented")
}

// UpdateRolloutViewer implements dao.RolloutViewer.
func (*RolloutViewer) UpdateRolloutViewer(ct context.Context, viewer entity.RolloutViewer) *errs.Error {
	panic("unimplemented")
}

var _ dao.RolloutViewer = (*RolloutViewer)(nil)

func NewRolloutViewer() *RolloutViewer {
	return &RolloutViewer{}
}
