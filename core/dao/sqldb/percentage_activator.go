package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/entity"
)

const percentageActivatorDaoName = "PercentageActivator"

type PercentageActivator struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.PercentageActivator = (*PercentageActivator)(nil)

func (p *PercentageActivator) FindPercentageActivatorByIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
) (entity.PartialPercentageActivator, *errs.Error) {
	p.metrics.ReportDaoOperation(percentageActivatorDaoName, "FindPercentageActivatorByIDWithTx")
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
	p.metrics.ReportDaoOperation(percentageActivatorDaoName, "FindPercentageActivatorByID")
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

func (p *PercentageActivator) CreatePercentageActivator(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
	activator entity.PartialPercentageActivator,
) *errs.Error {
	p.metrics.ReportDaoOperation(percentageActivatorDaoName, "CreatePercentageActivator")
	_, err := tx.SQLTx().ExecContext(ct, `
		INSERT INTO percentage_activator (
		    activator_id,
		    percentage
		)
		VALUES ($1, $2)
	`,
		activatorID,
		activator.Percentage)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (p *PercentageActivator) UpdatePercentageActivator(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
	activator entity.PartialPercentageActivator,
) *errs.Error {
	p.metrics.ReportDaoOperation(percentageActivatorDaoName, "UpdatePercentageActivator")
	_, err := tx.SQLTx().ExecContext(ct, `
		UPDATE percentage_activator
		SET percentage = $2
		WHERE activator_id = $1
	`,
		activatorID,
		activator.Percentage)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (p *PercentageActivator) DeletePercentageActivator(
	ct context.Context,
	tx *transaction.Transaction,
	activatorID uint64,
) *errs.Error {
	p.metrics.ReportDaoOperation(percentageActivatorDaoName, "DeletePercentageActivator")
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
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *PercentageActivator {
	return &PercentageActivator{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
