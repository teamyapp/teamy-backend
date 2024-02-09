package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type VersionSelector struct {
	transactionFactory transaction.Factory
}

var _ dao.VersionSelector = (*VersionSelector)(nil)

func (*VersionSelector) FindVersionSelectorByIDWithTx(ct context.Context, tx *transaction.Transaction, selectorID uint64) (entity.VersionSelector, *errs.Error) {
	versionSelector := entity.VersionSelector{}
	err := tx.SQLTx().QueryRowContext(ct,
		`
		SELECT 
			id, 
			type, 
			created_at, 
			updated_at
		FROM version_selector
		WHERE id = $1
		`,
		selectorID,
	).Scan(
		&versionSelector.ID,
		&versionSelector.Type,
		&versionSelector.CreatedAt,
		&versionSelector.UpdatedAt,
	)

	if err != nil {
		return entity.VersionSelector{}, errs.NewError(errs.Unknown, err.Error())
	}

	return versionSelector, nil
}

func (v *VersionSelector) FindVersionSelectorByID(ct context.Context, selectorID uint64) (entity.VersionSelector, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}

	tx, err := v.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.VersionSelector{}, err
	}

	defer tx.Rollback()
	return v.FindVersionSelectorByIDWithTx(ct, tx, selectorID)
}

func (*VersionSelector) CreateVersionSelector(
	ct context.Context,
	tx *transaction.Transaction,
	versionSelector entity.VersionSelector,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`
		INSERT INTO version_selector (
			id,
			type,
			created_at
		) VALUES (
			$1,
			$2,
			$3
		)
		`,
		versionSelector.ID,
		versionSelector.Type,
		versionSelector.CreatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (*VersionSelector) UpdateVersionSelector(ct context.Context, tx *transaction.Transaction, versionSelector entity.VersionSelector) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`
		UPDATE version_selector
		SET 
			type = $1,
			updated_at = $2
		WHERE id = $3
		`,
		versionSelector.Type,
		versionSelector.UpdatedAt,
		versionSelector.ID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (*VersionSelector) DeleteVersionSelector(ct context.Context, tx *transaction.Transaction, versionSelectorID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`
		DELETE FROM version_selector
		WHERE id = $1
		`,
		versionSelectorID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewVersionSelector(
	transactionFactory transaction.Factory,
) *VersionSelector {
	return &VersionSelector{
		transactionFactory: transactionFactory,
	}
}
