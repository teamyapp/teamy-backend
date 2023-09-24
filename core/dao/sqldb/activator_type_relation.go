package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type ActivatorTypeRelation struct {
	transactionFactory transaction.Factory
}

var _ dao.ActivatorTypeRelation = (*ActivatorTypeRelation)(nil)

func (a *ActivatorTypeRelation) FindActivatorTypeByIDWithTx(ct context.Context, tx *transaction.Transaction, activatorID uint64) (entity.ActivatorType, *errs.Error) {
	activatorTypeRelation := entity.ActivatorTypeRelation{}
	err := tx.SQLTx().QueryRowContext(ct,
		`
		    SELECT 
			  activator_id,
			  activator_type 
			FROM activator_type_relations 
			WHERE activator_id = $1`,
		activatorID,
	).Scan(
		&activatorTypeRelation.ActivatorType,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return "", errs.NewError(
			errs.NotFound,
			fmt.Sprintf(
				"activatorTypeRelation not found: activatorID=%v", activatorID))
	}

	if err != nil {
		return "", errs.NewError(errs.Unknown, err.Error())
	}

	return activatorTypeRelation.ActivatorType, nil
}

func (a *ActivatorTypeRelation) FindActivatorTypeByID(ct context.Context, activatorID uint64) (entity.ActivatorType, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return "", err
	}

	defer tx.Rollback()
	return a.FindActivatorTypeByIDWithTx(ct, tx, activatorID)
}

func (*ActivatorTypeRelation) CreateActivatorTypeRelation(ct context.Context, tx *transaction.Transaction, activatorTypeRelation entity.ActivatorTypeRelation) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		INSERT INTO activator_type_relations (
			activator_id,
			activator_type
		) 
		VALUES (
			$1,
			$2
		)`,
		activatorTypeRelation.ActivatorID,
		activatorTypeRelation.ActivatorType,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewActivatorTypeRelation(transactionFactory transaction.Factory) *ActivatorTypeRelation {
	return &ActivatorTypeRelation{transactionFactory: transactionFactory}
}
