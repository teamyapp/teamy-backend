package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type PercentageActivator struct {
	transactionFactory transaction.Factory
}

var _ dao.PercentageActivator = (*PercentageActivator)(nil)

func (*PercentageActivator) FindPercentageActivatorByIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	ActivatorID uint64,
) (entity.PercentageActivator, *errs.Error) {
	percentageActivator := entity.PercentageActivator{}
	err := tx.SQLTx().QueryRowContext(ct, `
		SELECT activator_id, percentage
		FROM percentage_activator
		WHERE activator_id = $1
	`, ActivatorID).Scan(
		&percentageActivator.Activator.ID,
		&percentageActivator.Percentage,
	)
	if err != nil {
		return percentageActivator, errs.NewError(errs.Unknown, err.Error())
	}

	return percentageActivator, nil
}

func (p *PercentageActivator) FindPercentageActivatorByID(ct context.Context, ActivatorID uint64) (entity.PercentageActivator, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := p.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.PercentageActivator{}, err
	}

	defer tx.Rollback()
	return p.FindPercentageActivatorByIDWithTx(ct, tx, ActivatorID)
}

func (*PercentageActivator) CreatePercentageActivator(
	ct context.Context,
	tx *transaction.Transaction,
	activator entity.PercentageActivator,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		INSERT INTO percentage_activator (
		    activator_id, 
		    percentage
		)
		VALUES ($1, $2)
	`, activator.Activator.ID, activator.Percentage)
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
