package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const activatorDaoName = "Activator"

type Activator struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

func (a *Activator) CreateActivator(ct context.Context, tx *transaction.Transaction, activator entity.Activator) *errs.Error {
	a.metrics.ReportDaoOperation(activatorDaoName, "CreateActivator")
	_, err := tx.SQLTx().ExecContext(ct, `
		INSERT INTO activator (
			id,
			type,
			locked,
			created_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4
		)`,
		activator.ID,
		activator.Type,
		activator.Locked,
		activator.CreatedAt,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a *Activator) FindActivatorByIDWithTx(ct context.Context, tx *transaction.Transaction, activatorID uint64) (entity.Activator, *errs.Error) {
	a.metrics.ReportDaoOperation(activatorDaoName, "FindActivatorByIDWithTx")
	activator := entity.Activator{}
	err := tx.SQLTx().QueryRowContext(ct, `
		SELECT
			id,
			type,
			locked,
			created_at,
			updated_at
		FROM activator
		WHERE id = $1`,
		activatorID,
	).Scan(
		&activator.ID,
		&activator.Type,
		&activator.Locked,
		&activator.CreatedAt,
		&activator.UpdatedAt,
	)

	if err != nil {
		return entity.Activator{}, errs.NewError(errs.Unknown, err.Error())
	}

	return activator, nil
}

func (a *Activator) FindActivatorByID(ct context.Context, activatorID uint64) (entity.Activator, *errs.Error) {
	a.metrics.ReportDaoOperation(activatorDaoName, "FindActivatorByID")
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.Activator{}, err
	}

	defer tx.Rollback()
	return a.FindActivatorByIDWithTx(ct, tx, activatorID)
}

func (a *Activator) UpdateActivator(ct context.Context, tx *transaction.Transaction, activator entity.Activator) *errs.Error {
	a.metrics.ReportDaoOperation(activatorDaoName, "UpdateActivator")
	_, err := tx.SQLTx().ExecContext(ct, `
		UPDATE activator
		SET
			type = $1,
			locked = $2,
			updated_at = $3
		WHERE id = $4`,
		activator.Type,
		activator.Locked,
		activator.UpdatedAt,
		activator.ID,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a *Activator) DeleteActivator(ct context.Context, tx *transaction.Transaction, activatorID uint64) *errs.Error {
	a.metrics.ReportDaoOperation(activatorDaoName, "DeleteActivator")
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM activator
		WHERE id = $1`,
		activatorID,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

var _ dao.Activator = (*Activator)(nil)

func NewActivator(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *Activator {
	return &Activator{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
