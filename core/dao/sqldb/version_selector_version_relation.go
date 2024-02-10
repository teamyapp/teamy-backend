package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type VersionSelectorVersionRelation struct {
	transactionFactory transaction.Factory
}

var _ dao.VersionSelectorVersionRelation = (*VersionSelectorVersionRelation)(nil)

func (*VersionSelectorVersionRelation) FindVersionNumbersBySelectorIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	selectorID uint64,
) ([]int, *errs.Error) {
	versionNumbers := make([]int, 0)
	rows, err := tx.SQLTx().QueryContext(ct,
		`
		SELECT version_number
		FROM version_selector_version_relation
		WHERE version_selector_id = $1
		`,
		selectorID,
	)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		var versionNumber int
		err := rows.Scan(&versionNumber)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		versionNumbers = append(versionNumbers, versionNumber)
	}

	return versionNumbers, nil
}

func (v *VersionSelectorVersionRelation) FindVersionNumbersBySelectorID(ct context.Context, selectorID uint64) ([]int, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}

	tx, err := v.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return v.FindVersionNumbersBySelectorIDWithTx(ct, tx, selectorID)
}

func (*VersionSelectorVersionRelation) CreateVersionSelectorVersionRelation(
	ct context.Context,
	tx *transaction.Transaction,
	versionSelectorVersionRelation entity.VersionSelectorVersionRelation,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`
		INSERT INTO version_selector_version_relation (
			version_selector_id,
			version_number
		) 
		VALUES (
			$1,
			$2
		);
		`,
		versionSelectorVersionRelation.VersionSelectorID,
		versionSelectorVersionRelation.VersionNumber,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (*VersionSelectorVersionRelation) UpdateVersionSelectorVersionRelation(
	ct context.Context,
	tx *transaction.Transaction,
	versionSelectorID uint64,
	versionNumber int,
	versionSelectorVersionRelation entity.VersionSelectorVersionRelation,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`
		UPDATE version_selector_version_relation
		SET version_number = $1
		WHERE version_number = $2
		AND version_selector_id = $3
		`,
		versionSelectorVersionRelation.VersionNumber,
		versionNumber,
		versionSelectorVersionRelation.VersionSelectorID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (v *VersionSelectorVersionRelation) DeleteVersionSelectorVersionRelationBySelectorID(
	ct context.Context,
	tx *transaction.Transaction,
	versionSelectorID uint64,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`
		DELETE FROM version_selector_version_relation
		WHERE version_selector_id = $1
		`,
		versionSelectorID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (v *VersionSelectorVersionRelation) DeleteVersionSelectorVersionRelationBySelectorIDAndVersionNumber(
	ct context.Context,
	tx *transaction.Transaction,
	versionSelectorID uint64,
	versionNumber int,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`
		DELETE FROM version_selector_version_relation
		WHERE version_selector_id = $1 AND version_number = $2
		`,
		versionSelectorID,
		versionNumber,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewVersionSelectorVersionRelation(
	transactionFactory transaction.Factory,
) *VersionSelectorVersionRelation {
	return &VersionSelectorVersionRelation{
		transactionFactory: transactionFactory,
	}
}
