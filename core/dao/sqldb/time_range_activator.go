package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TimeRangeActivator struct {
	transactionFactory transaction.Factory
}

var _ dao.TimeRangeActivator = (*TimeRangeActivator)(nil)

func (*TimeRangeActivator) FindTimeRangeActivatorByIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
) (entity.TimeRangeActivator, *errs.Error) {
	timeRangeActivator := entity.TimeRangeActivator{}
	err := tx.SQLTx().QueryRowContext(
		ct,
		`
		SELECT
			activator_id,
			start_time,
			end_time,
		FROM time_range_activator
		WHERE activator_id = $1;`,
		activatorID,
	).Scan(
		&timeRangeActivator.Activator.ID,
		&timeRangeActivator.StartAt,
		&timeRangeActivator.EndAt,
	)

	if err != nil {
		return timeRangeActivator, errs.NewError(errs.Unknown, err.Error())
	}

	return timeRangeActivator, nil
}

func (t *TimeRangeActivator) FindTimeRangeActivatorByID(ct context.Context, activatorID uint64) (entity.TimeRangeActivator, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := t.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.TimeRangeActivator{}, err
	}

	defer tx.Rollback()
	return t.FindTimeRangeActivatorByIDWithTx(ct, tx, activatorID)
}

func (*TimeRangeActivator) CreateTimeRangeActivator(
	ct context.Context,
	tx *transaction.Transaction,
	activator entity.TimeRangeActivator,
) *errs.Error {
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
		activator.Activator.ID,
		activator.StartAt,
		activator.EndAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTimeRangeActivator(
	transactionFactory transaction.Factory,
) *TimeRangeActivator {
	return &TimeRangeActivator{
		transactionFactory: transactionFactory,
	}
}
