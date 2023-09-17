package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type MaxViewersActivator struct {
	transactionFactory transaction.Factory
}

var _ dao.MaxViewersActivator = (*MaxViewersActivator)(nil)

func (*MaxViewersActivator) FindMaxViewersActivatorByIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	ActivatorID uint64,
) (entity.MaxViewersActivator, *errs.Error) {
	maxViewersActivator := entity.MaxViewersActivator{}
	err := tx.SQLTx().QueryRowContext(ct,
		`
			SELECT
				activator_id,
				max_viewers
			FROM max_viewers_activator
			WHERE activator_id = $1
		`,
		ActivatorID,
	).Scan(
		&maxViewersActivator.Activator.ID,
		&maxViewersActivator.MaxViewers,
	)

	if err != nil {
		return entity.MaxViewersActivator{}, errs.NewError(errs.Unknown, err.Error())
	}

	return maxViewersActivator, nil
}

func (m *MaxViewersActivator) FindMaxViewersActivatorByID(ct context.Context, ActivatorID uint64) (entity.MaxViewersActivator, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}

	tx, err := m.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.MaxViewersActivator{}, err
	}

	defer tx.Rollback()
	return m.FindMaxViewersActivatorByIDWithTx(ct, tx, ActivatorID)
}

func (*MaxViewersActivator) CreateMaxViewersActivator(
	ct context.Context,
	tx *transaction.Transaction,
	activator entity.MaxViewersActivator,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
			INSERT INTO max_viewers_activator (
				activator_id,
				max_viewers
			) VALUES (
				$1,
				$2
			)
		`,
		activator.Activator.ID,
		activator.MaxViewers,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewMaxViewersActivator(transactionFactory transaction.Factory) *MaxViewersActivator {
	return &MaxViewersActivator{
		transactionFactory: transactionFactory,
	}
}
