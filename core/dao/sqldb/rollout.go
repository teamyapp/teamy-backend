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

type Rollout struct {
	transactionFactory transaction.Factory
}

var _ dao.Rollout = (*Rollout)(nil)

func (*Rollout) FindRolloutByIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	rolloutID uint64,
) (entity.Rollout, *errs.Error) {
	rollout := entity.Rollout{}
	err := tx.SQLTx().QueryRowContext(ct,
		`
		SELECT
			id,
			activator_id,
			version_selector_id,
			viewers,
			is_enabled,
			created_at,
			updated_at
		FROM rollout
		WHERE id = $1
		`,
		rolloutID,
	).Scan(
		&rollout.ID,
		&rollout.ActivatorID,
		&rollout.SelectorID,
		&rollout.Viewers,
		&rollout.IsEnabled,
		&rollout.CreatedAt,
		&rollout.UpdatedAt,
	)

	if err != nil {
		return rollout, errs.NewError(errs.Unknown, err.Error())
	}

	return rollout, nil
}

func (r *Rollout) FindRolloutByID(ct context.Context, rolloutIDs uint64) (entity.Rollout, *errs.Error) {
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

func (*Rollout) FindRolloutsByIDsWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	rolloutIDs []uint64,
) ([]entity.Rollout, *errs.Error) {
	if len(rolloutIDs) == 0 {
		return nil, errs.NewError(errs.InvalidArgument, "rolloutIDs is empty")
	}

	rollouts := make([]entity.Rollout, 0)
	idsString := toIDsString(rolloutIDs)
	query := fmt.Sprintf(
		`
		SELECT
			id,
			activator_id,
			version_selector_id,
			viewers,
			is_enabled,
			created_at,
			updated_at
		FROM rollout
		WHERE id IN (%s)
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
			&rollout.ActivatorID,
			&rollout.SelectorID,
			&rollout.Viewers,
			&rollout.IsEnabled,
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

func (*Rollout) CreateRollout(
	ct context.Context,
	tx *transaction.Transaction,
	rollout entity.Rollout,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`
		INSERT INTO rollout (
			id,
			activator_id,
			version_selector_id,
			viewers,
			is_enabled
		)
		VALUES ($1, $2, $3, $4, $5)
		`,
		rollout.ID,
		rollout.ActivatorID,
		rollout.SelectorID,
		rollout.Viewers,
		rollout.IsEnabled,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (*Rollout) UpdateRollout(
	ct context.Context,
	tx *transaction.Transaction,
	rollout entity.Rollout,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
		UPDATE rollout
		SET
			activator_id = $1,
			version_selector_id = $2,
			viewers = $3,
			is_enabled = $4,
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
		`,
		rollout.ActivatorID,
		rollout.SelectorID,
		rollout.Viewers,
		rollout.IsEnabled,
		rollout.ID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (*Rollout) DeleteRollout(
	ct context.Context,
	tx *transaction.Transaction,
	rolloutID uint64,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`
		DELETE FROM rollout
		WHERE id = $1
		`,
		rolloutID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewRollout(
	transactionFactory transaction.Factory,
) *Rollout {
	return &Rollout{
		transactionFactory: transactionFactory,
	}
}
