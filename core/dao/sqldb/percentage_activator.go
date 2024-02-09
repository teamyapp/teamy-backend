package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/entity"
)

type PercentageActivator struct {
	transactionFactory transaction.Factory
}

var _ dao.PercentageActivator = (*PercentageActivator)(nil)

func (*PercentageActivator) FindPercentageActivatorByIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
) (entity.PartialPercentageActivator, *errs.Error) {
	percentageActivator := entity.PartialPercentageActivator{}
	err := tx.SQLTx().QueryRowContext(ct, `
		SELECT percentage
		FROM percentage_activator
		WHERE activator_id = $1
	`, activatorID).Scan(
		&percentageActivator.Percentage,
	)
	if err != nil {
		return percentageActivator, errs.NewError(errs.Unknown, err.Error())
	}

	return percentageActivator, nil
}

func (p *PercentageActivator) FindPercentageActivatorByID(ct context.Context, ActivatorID uint64) (entity.PartialPercentageActivator, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := p.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.PartialPercentageActivator{}, err
	}

	defer tx.Rollback()
	return p.FindPercentageActivatorByIDWithTx(ct, tx, ActivatorID)
}

func (*PercentageActivator) CreatePercentageActivator(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
	activator entity.PartialPercentageActivator,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		INSERT INTO percentage_activator (
			activator_id,
		    percentage
		)
		VALUES ($1, $2)
	`, activatorID, activator.Percentage)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (*PercentageActivator) UpdatePercentageActivator(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
	activator entity.PartialPercentageActivator,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		UPDATE percentage_activator
		SET percentage = $2
		WHERE activator_id = $1
	`, activatorID, activator.Percentage)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (*PercentageActivator) DeletePercentageActivator(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM percentage_activator
		WHERE activator_id = $1
	`, activatorID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewPercentageActivator(
	transactionFactory transaction.Factory,
) *PercentageActivator {
	return &PercentageActivator{
		transactionFactory: transactionFactory,
	}
}
