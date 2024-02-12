package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Activator struct {
	transactionFactory transaction.Factory
}

func (*Activator) CreateActivator(ct context.Context, tx *transaction.Transaction, activator entity.Activator) *errs.Error {
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

func (*Activator) FindActivatorByIDWithTx(ct context.Context, tx *transaction.Transaction, activatorID uint64) (entity.Activator, *errs.Error) {
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

func (*Activator) UpdateActivator(ct context.Context, tx *transaction.Transaction, activator entity.Activator) *errs.Error {
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

func (*Activator) DeleteActivator(ct context.Context, tx *transaction.Transaction, activatorID uint64) *errs.Error {
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

func NewActivator(transactionFactory transaction.Factory) *Activator {
	return &Activator{
		transactionFactory: transactionFactory,
	}
}
