package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type RolloutViewer struct {
	transactionFactory transaction.Factory
}

var _ dao.RolloutViewer = (*RolloutViewer)(nil)

func (*RolloutViewer) FindRolloutViewerByViewerIDAndRolloutIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	viewerID uint64,
	rolloutID uint64,
) (entity.RolloutViewer, *errs.Error) {
	rolloutViewer := entity.RolloutViewer{}
	err := tx.SQLTx().QueryRowContext(
		ct,
		`
		SELECT
			rollout_id,
			viewer_id,
			version_number,
			is_activated,
			created_at,
			updated_at
		FROM rollout_viewer
		WHERE viewer_id = $1 AND rollout_id = $2;
		`,
		viewerID,
		rolloutID,
	).Scan(
		&rolloutViewer.RolloutID,
		&rolloutViewer.ViewerID,
		&rolloutViewer.VersionNumber,
		&rolloutViewer.IsActivated,
		&rolloutViewer.CreatedAt,
		&rolloutViewer.UpdatedAt,
	)

	if err != nil {
		return entity.RolloutViewer{}, errs.NewError(errs.Unknown, err.Error())
	}

	return rolloutViewer, nil
}

func (r *RolloutViewer) FindRolloutViewerByViewerIDAndRolloutID(ct context.Context, viewerID uint64, RolloutID uint64) (entity.RolloutViewer, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := r.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.RolloutViewer{}, err
	}

	defer tx.Rollback()
	return r.FindRolloutViewerByViewerIDAndRolloutIDWithTx(ct, tx, viewerID, RolloutID)
}

func (*RolloutViewer) CreateRolloutViewer(
	ct context.Context,
	tx *transaction.Transaction,
	viewer entity.RolloutViewer,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		INSERT INTO rollout_viewer (
			rollout_id,
			viewer_id,
			version_number,
			is_activated,
			created_at,
			updated_at
		) 
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6
		);
	`,
		viewer.RolloutID,
		viewer.ViewerID,
		viewer.VersionNumber,
		viewer.IsActivated,
		viewer.CreatedAt,
		viewer.UpdatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (*RolloutViewer) UpdateRolloutViewer(
	ct context.Context,
	tx *transaction.Transaction,
	viewer entity.RolloutViewer,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		UPDATE rollout_viewer 
		SET
			is_activated = $1,
			updated_at = $2
		WHERE viewer_id = $3 AND rollout_id = $4;
	`,
		viewer.IsActivated,
		viewer.UpdatedAt,
		viewer.ViewerID,
		viewer.RolloutID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (r *RolloutViewer) DeleteRolloutViewersByRolloutID(ct context.Context, tx *transaction.Transaction, rolloutID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM rollout_viewer
		WHERE rollout_id = $1;
	`,
		rolloutID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewRolloutViewer(
	transactionFactory transaction.Factory,
) *RolloutViewer {
	return &RolloutViewer{
		transactionFactory: transactionFactory,
	}
}
