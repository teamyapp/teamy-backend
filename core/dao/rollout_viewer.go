package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type RolloutViewer interface {
	FindRolloutViewerByViewerIDAndRolloutID(ct context.Context, viewerID uint64, RolloutID uint64) (entity.RolloutViewer, *errs.Error)
	UpdateRolloutViewer(ct context.Context, viewer entity.RolloutViewer) *errs.Error
	CreateRolloutViewer(ct context.Context, viewer entity.RolloutViewer) (entity.RolloutViewer, *errs.Error)
}
