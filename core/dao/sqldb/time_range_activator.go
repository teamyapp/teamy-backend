package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/entity"
)

const timeRangeActivatorDaoName = "TimeRangeActivator"

type TimeRangeActivator struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.TimeRangeActivator = (*TimeRangeActivator)(nil)

func (t *TimeRangeActivator) FindTimeRangeActivatorByIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
) (entity.PartialTimeRangeActivator, *errs.Error) {
	t.metrics.ReportDaoOperation(timeRangeActivatorDaoName, "FindTimeRangeActivatorByIDWithTx")
	timeRangeActivator := entity.PartialTimeRangeActivator{}
	err := tx.SQLTx().QueryRowContext(
		ct,
		`
		SELECT
			start_time,
			end_time
		FROM time_range_activator
		WHERE activator_id = $1;`,
		activatorID,
	).Scan(
		&timeRangeActivator.StartAt,
		&timeRangeActivator.EndAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return timeRangeActivator, errs.NewError(errs.NotFound, fmt.Sprintf("time range activator not found with id %d", activatorID))
		}

		return timeRangeActivator, errs.NewError(errs.Unknown, err.Error())
	}

	return timeRangeActivator, nil
}

func (t *TimeRangeActivator) FindTimeRangeActivatorByID(ct context.Context, activatorID uint64) (entity.PartialTimeRangeActivator, *errs.Error) {
	t.metrics.ReportDaoOperation(timeRangeActivatorDaoName, "FindTimeRangeActivatorByID")
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := t.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.PartialTimeRangeActivator{}, err
	}

	defer tx.Rollback()
	return t.FindTimeRangeActivatorByIDWithTx(ct, tx, activatorID)
}

func (t *TimeRangeActivator) CreateTimeRangeActivator(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
	activator entity.PartialTimeRangeActivator,
) *errs.Error {
	t.metrics.ReportDaoOperation(timeRangeActivatorDaoName, "CreateTimeRangeActivator")
	_, err := tx.SQLTx().ExecContext(
		ct,
		`INSERT INTO time_range_activator (
			activator_id,
			start_time,
			end_time
		)
		VALUES (
			$1,
			$2,
			$3
		);`,
		activatorID,
		activator.StartAt,
		activator.EndAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t *TimeRangeActivator) UpdateTimeRangeActivator(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
	activator entity.PartialTimeRangeActivator) *errs.Error {
	t.metrics.ReportDaoOperation(timeRangeActivatorDaoName, "UpdateTimeRangeActivator")
	_, err := tx.SQLTx().ExecContext(
		ct,
		`UPDATE time_range_activator
		SET
			start_time = $2,
			end_time = $3
		WHERE activator_id = $1;`,
		activatorID,
		activator.StartAt,
		activator.EndAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t *TimeRangeActivator) DeleteTimeRangeActivator(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
) *errs.Error {
	t.metrics.ReportDaoOperation(timeRangeActivatorDaoName, "DeleteTimeRangeActivator")
	_, err := tx.SQLTx().ExecContext(
		ct,
		`DELETE FROM time_range_activator
		WHERE activator_id = $1;`,
		activatorID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTimeRangeActivator(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *TimeRangeActivator {
	return &TimeRangeActivator{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
