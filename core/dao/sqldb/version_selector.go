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
		SELECT id, type
		FROM version_selector
		WHERE id = $1
		`,
		selectorID,
	).Scan(
		&versionSelector.ID,
		&versionSelector.Type,
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
			type
		) VALUES (
			$1,
			$2
		)
		`,
		versionSelector.ID,
		versionSelector.Type,
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
