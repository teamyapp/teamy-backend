package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/entity"
)

const maxViewersActivatorDaoName = "MaxViewersActivator"

type MaxViewersActivator struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.MaxViewersActivator = (*MaxViewersActivator)(nil)

func (m *MaxViewersActivator) FindMaxViewersActivatorByIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
) (entity.PartialMaxViewersActivator, *errs.Error) {
	m.metrics.ReportDaoOperation(maxViewersActivatorDaoName, "FindMaxViewersActivatorByIDWithTx")
	maxViewersActivator := entity.PartialMaxViewersActivator{}
	err := tx.SQLTx().QueryRowContext(ct,
		`
			SELECT
				max_viewers
			FROM max_viewers_activator
			WHERE activator_id = $1
		`,
		activatorID,
	).Scan(
		&maxViewersActivator.MaxViewers,
	)

	if err != nil {
		return entity.PartialMaxViewersActivator{}, errs.NewError(errs.Unknown, err.Error())
	}

	return maxViewersActivator, nil
}

func (m *MaxViewersActivator) FindMaxViewersActivatorByID(ct context.Context, ActivatorID uint64) (entity.PartialMaxViewersActivator, *errs.Error) {
	m.metrics.ReportDaoOperation(maxViewersActivatorDaoName, "FindMaxViewersActivatorByID")
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := m.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.PartialMaxViewersActivator{}, err
	}

	defer tx.Rollback()
	return m.FindMaxViewersActivatorByIDWithTx(ct, tx, ActivatorID)
}

func (m *MaxViewersActivator) CreateMaxViewersActivator(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
	activator entity.PartialMaxViewersActivator,
) *errs.Error {
	m.metrics.ReportDaoOperation(maxViewersActivatorDaoName, "CreateMaxViewersActivator")
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
			INSERT INTO max_viewers_activator (
				activator_id,
				max_viewers
			)
			VALUES (
				$1,
				$2
			)
		`,
		activatorID,
		activator.MaxViewers,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (m *MaxViewersActivator) UpdateMaxViewersActivator(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
	activator entity.PartialMaxViewersActivator,
) *errs.Error {
	m.metrics.ReportDaoOperation(maxViewersActivatorDaoName, "UpdateMaxViewersActivator")
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
			UPDATE max_viewers_activator
			SET max_viewers = $1
			WHERE activator_id = $2
		`,
		activator.MaxViewers,
		activatorID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (m *MaxViewersActivator) DeleteMaxViewersActivator(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
) *errs.Error {
	m.metrics.ReportDaoOperation(maxViewersActivatorDaoName, "DeleteMaxViewersActivator")
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
			DELETE FROM max_viewers_activator
			WHERE activator_id = $1
		`,
		activatorID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewMaxViewersActivator(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *MaxViewersActivator {
	return &MaxViewersActivator{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
