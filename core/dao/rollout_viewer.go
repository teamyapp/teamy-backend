package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type RolloutViewer interface {
	FindRolloutViewerByViewerIDAndRolloutIDWithTx(ct context.Context, tx *transaction.Transaction, viewerID uint64, rolloutID uint64) (entity.RolloutViewer, *errs.Error)
	FindRolloutViewerByViewerIDAndRolloutID(ct context.Context, viewerID uint64, RolloutID uint64) (entity.RolloutViewer, *errs.Error)
	UpdateRolloutViewer(ct context.Context, tx *transaction.Transaction, viewer entity.RolloutViewer) *errs.Error
	CreateRolloutViewer(ct context.Context, tx *transaction.Transaction, viewer entity.RolloutViewer) *errs.Error
}
