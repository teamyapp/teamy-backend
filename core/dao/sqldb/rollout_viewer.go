package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type RolloutViewer struct{}

var _ dao.RolloutViewer = (*RolloutViewer)(nil)

func (*RolloutViewer) FindRolloutViewerByViewerIDAndRolloutID(ct context.Context, viewerID uint64, RolloutID uint64) (entity.RolloutViewer, *errs.Error) {
	panic("unimplemented")
}

func (*RolloutViewer) CreateRolloutViewer(ct context.Context, viewer entity.RolloutViewer) (entity.RolloutViewer, *errs.Error) {
	panic("unimplemented")
}

func (*RolloutViewer) UpdateRolloutViewer(ct context.Context, viewer entity.RolloutViewer) *errs.Error {
	panic("unimplemented")
}

func NewRolloutViewer() *RolloutViewer {
	return &RolloutViewer{}
}
