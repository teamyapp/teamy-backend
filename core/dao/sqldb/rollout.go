package sqldb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const rolloutDaoName = "Rollout"

type Rollout struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.Rollout = (*Rollout)(nil)

func (r *Rollout) FindRolloutByIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	rolloutID uint64,
) (entity.Rollout, *errs.Error) {
	r.metrics.ReportDaoOperation(rolloutDaoName, "FindRolloutByIDWithTx")
	rollout := entity.Rollout{}
	err := tx.SQLTx().QueryRowContext(ct,
		`
		SELECT
			id,
			name,
			activator_id,
			version_selector_id,
			viewers,
			is_enabled,
			locked,
			created_at,
			updated_at
		FROM rollout
		WHERE id = $1;
		`,
		rolloutID,
	).Scan(
		&rollout.ID,
		&rollout.Name,
		&rollout.ActivatorID,
		&rollout.SelectorID,
		&rollout.Viewers,
		&rollout.IsEnabled,
		&rollout.Locked,
		&rollout.CreatedAt,
		&rollout.UpdatedAt,
	)

	if err != nil {
		return rollout, errs.NewError(errs.Unknown, err.Error())
	}

	return rollout, nil
}

func (r *Rollout) FindRolloutByID(ct context.Context, rolloutIDs uint64) (entity.Rollout, *errs.Error) {
	r.metrics.ReportDaoOperation(rolloutDaoName, "FindRolloutByID")
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := r.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.Rollout{}, err
	}

	defer tx.Rollback()
	return r.FindRolloutByIDWithTx(ct, tx, rolloutIDs)
}

func (r *Rollout) FindRolloutsByIDsWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	rolloutIDs []uint64,
) ([]entity.Rollout, *errs.Error) {
	r.metrics.ReportDaoOperation(rolloutDaoName, "FindRolloutsByIDsWithTx")
	if len(rolloutIDs) == 0 {
		return nil, nil
	}

	rollouts := make([]entity.Rollout, 0)
	idsString := toIDsString(rolloutIDs)
	query := fmt.Sprintf(
		`
		SELECT
			id,
			name,
			activator_id,
			version_selector_id,
			viewers,
			is_enabled,
			locked,
			created_at,
			updated_at
		FROM rollout
		WHERE id IN (%s);
		`,
		idsString,
	)

	rows, err := tx.SQLTx().QueryContext(ct, query)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		var rollout entity.Rollout
		err := rows.Scan(
			&rollout.ID,
			&rollout.Name,
			&rollout.ActivatorID,
			&rollout.SelectorID,
			&rollout.Viewers,
			&rollout.IsEnabled,
			&rollout.Locked,
			&rollout.CreatedAt,
			&rollout.UpdatedAt,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		rollouts = append(rollouts, rollout)
	}

	return rollouts, nil
}

func (r *Rollout) FindRolloutsByIDs(ct context.Context, rolloutIDs []uint64) ([]entity.Rollout, *errs.Error) {
	r.metrics.ReportDaoOperation(rolloutDaoName, "FindRolloutsByIDs")
	opt := sql.TxOptions{
		ReadOnly: true,
	}

	tx, err := r.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return r.FindRolloutsByIDsWithTx(ct, tx, rolloutIDs)
}

func (r *Rollout) CreateRollout(
	ct context.Context,
	tx *transaction.Transaction,
	rollout entity.Rollout,
) *errs.Error {
	r.metrics.ReportDaoOperation(rolloutDaoName, "CreateRollout")
	_, err := tx.SQLTx().ExecContext(ct,
		`
		INSERT INTO rollout (
			id,
			name,
			activator_id,
			version_selector_id,
			viewers,
			is_enabled,
			locked,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
		`,
		rollout.ID,
		rollout.Name,
		rollout.ActivatorID,
		rollout.SelectorID,
		rollout.Viewers,
		rollout.IsEnabled,
		rollout.Locked,
		rollout.CreatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (r *Rollout) UpdateRollout(
	ct context.Context,
	tx *transaction.Transaction,
	rollout entity.Rollout,
) *errs.Error {
	r.metrics.ReportDaoOperation(rolloutDaoName, "UpdateRollout")
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
		UPDATE rollout
		SET
			activator_id = $1,
			name = $2,
			version_selector_id = $3,
			viewers = $4,
			is_enabled = $5,
			locked = $6,
			updated_at = $7,
			created_at = $8
		WHERE id = $9;
		`,
		rollout.ActivatorID,
		rollout.Name,
		rollout.SelectorID,
		rollout.Viewers,
		rollout.IsEnabled,
		rollout.Locked,
		rollout.UpdatedAt,
		rollout.CreatedAt,
		rollout.ID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (r *Rollout) DeleteRollout(
	ct context.Context,
	tx *transaction.Transaction,
	rolloutID uint64,
) *errs.Error {
	r.metrics.ReportDaoOperation(rolloutDaoName, "DeleteRollout")
	_, err := tx.SQLTx().ExecContext(ct,
		`
		DELETE FROM rollout
		WHERE id = $1;
		`,
		rolloutID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewRollout(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *Rollout {
	return &Rollout{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
