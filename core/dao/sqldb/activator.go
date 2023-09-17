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

// CREATE TABLE activator (
//     id BIGINT PRIMARY KEY,
//     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     updated_at  TIMESTAMP
// );

func (*Activator) CreateActivator(ct context.Context, tx *transaction.Transaction, activator entity.Activator) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		INSERT INTO activator (
			id,
			created_at,
			updated_at
		) VALUES (
			$1
		)`,
		activator.ID,
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
			created_at,
			updated_at
			FROM activator
			WHERE id = $1`,
		activatorID,
	).Scan(
		&activator.ID,
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
			updated_at = $2
		WHERE id = $1`,
		activator.ID,
		activator.UpdatedAt,
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
