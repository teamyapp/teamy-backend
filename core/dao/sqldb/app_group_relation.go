package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppGroupRelation struct {
	transactionFactory transaction.Factory
}

var _ dao.AppGroupRelation = (*AppGroupRelation)(nil)

func (*AppGroupRelation) FindAppGroupRelationTypeWithTx(ct context.Context, tx *transaction.Transaction, appID uint64, groupID uint64) (entity.AppGroupRelationType, *errs.Error) {
	var appGroupRelationType entity.AppGroupRelationType
	err := tx.SQLTx().QueryRowContext(ct, `
		SELECT type
		FROM app_group_relation
		WHERE app_id = $1 AND group_id = $2`,
		appID,
		groupID,
	).Scan(
		&appGroupRelationType,
	)

	if err != nil {
		return "", errs.NewError(errs.Unknown, err.Error())
	}

	return appGroupRelationType, nil
}

func (a *AppGroupRelation) FindAppGroupRelationType(ct context.Context, appID uint64, groupID uint64) (entity.AppGroupRelationType, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return "", err
	}

	defer tx.Rollback()
	return a.FindAppGroupRelationTypeWithTx(ct, tx, appID, groupID)
}

func (a *AppGroupRelation) FindAppIDByGroupIDWithTx(ct context.Context, tx *transaction.Transaction, groupID uint64) (uint64, *errs.Error) {
	var appID uint64
	err := tx.SQLTx().QueryRowContext(ct,
		`
		    SELECT 
			   app_id
			FROM app_group_relation
			WHERE group_id = $1`,
		groupID,
	).Scan(
		&appID,
	)

	if err != nil {
		return 0, errs.NewError(errs.Unknown, err.Error())
	}

	return appID, nil
}

func (a *AppGroupRelation) FindAppIDByGroupID(ct context.Context, groupID uint64) (uint64, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return 0, err
	}

	defer tx.Rollback()
	return a.FindAppIDByGroupIDWithTx(ct, tx, groupID)
}

func (a *AppGroupRelation) FindGroupIDsByAppIDAndRelationTypeWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	appID uint64,
	appGroupRelationType entity.AppGroupRelationType,
) ([]uint64, *errs.Error) {
	rows, err := tx.SQLTx().QueryContext(ct,
		`
		SELECT
			group_id
		FROM app_group_relation
		WHERE app_id = $1 AND type = $2`,
		appID,
		appGroupRelationType,
	)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	var groupIDs []uint64
	for rows.Next() {
		var groupID uint64
		err := rows.Scan(&groupID)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		groupIDs = append(groupIDs, groupID)
	}

	return groupIDs, nil
}

func (a *AppGroupRelation) FindGroupIDsByAppIDAndRelationType(ct context.Context, appID uint64, appGroupRelationType entity.AppGroupRelationType) ([]uint64, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return a.FindGroupIDsByAppIDAndRelationTypeWithTx(ct, tx, appID, appGroupRelationType)
}

func (*AppGroupRelation) CreateAppGroupRelation(ct context.Context, tx *transaction.Transaction, appGroupRelation entity.AppGroupRelation) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`INSERT INTO app_group_relation (
			app_id,
			group_id,
			type
		) 
		VALUES (
			$1,
			$2,
			$3
		)`,
		appGroupRelation.AppID,
		appGroupRelation.GroupID,
		appGroupRelation.Type,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (*AppGroupRelation) DeleteAppGroupRelation(ct context.Context, tx *transaction.Transaction, appID uint64, groupID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`DELETE FROM app_group_relation WHERE app_id = $1 AND group_id = $2`,
		appID,
		groupID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewAppGroupRelation(transactionFactory transaction.Factory) *AppGroupRelation {
	return &AppGroupRelation{
		transactionFactory: transactionFactory,
	}
}
