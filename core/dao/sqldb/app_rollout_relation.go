package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppRolloutRelation struct {
	transactionFactory transaction.Factory
}

var _ dao.AppRolloutRelation = (*AppRolloutRelation)(nil)

func (a *AppRolloutRelation) FindRolloutIDsByAppIDAndRelationTypeWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	appID uint64,
	rolloutType entity.AppRolloutRelationType,
) ([]uint64, *errs.Error) {
	rows, err := tx.SQLTx().QueryContext(ct,
		`
		SELECT
			rollout_id
		FROM app_rollout_relation
		WHERE app_id = $1 AND type = $2`,
		appID,
		rolloutType,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	var rolloutIDs []uint64
	for rows.Next() {
		var rolloutID uint64
		err := rows.Scan(&rolloutID)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		rolloutIDs = append(rolloutIDs, rolloutID)
	}

	return rolloutIDs, nil
}

func (a *AppRolloutRelation) FindAppRolloutByAppIDAndRolloutIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	appID, rolloutID uint64,
) (*entity.AppRolloutRelation, *errs.Error) {
	row := tx.SQLTx().QueryRowContext(ct,
		`
		SELECT
			app_id,
			rollout_id,
			type
		FROM app_rollout_relation
		WHERE app_id = $1 AND rollout_id = $2`,
		appID,
		rolloutID,
	)

	var appRolloutRelation entity.AppRolloutRelation
	err := row.Scan(
		&appRolloutRelation.AppID,
		&appRolloutRelation.RolloutID,
		&appRolloutRelation.Type,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.NewError(errs.NotFound, "app_rollout_relation not found")
		}

		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	return &appRolloutRelation, nil
}

func (a *AppRolloutRelation) FindRolloutIDsByAppIDAndRelationType(ct context.Context, appID uint64, rolloutType entity.AppRolloutRelationType) ([]uint64, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}

	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return a.FindRolloutIDsByAppIDAndRelationTypeWithTx(ct, tx, appID, rolloutType)
}

func (*AppRolloutRelation) CreateAppRolloutRelation(ct context.Context, tx *transaction.Transaction, appRolloutRelation entity.AppRolloutRelation) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`INSERT INTO app_rollout_relation (
			app_id,
			rollout_id,
			type
		)
		VALUES (
			$1,
			$2,
			$3
		)`,
		appRolloutRelation.AppID,
		appRolloutRelation.RolloutID,
		appRolloutRelation.Type,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewAppRolloutRelation(transactionFactory transaction.Factory) *AppRolloutRelation {
	return &AppRolloutRelation{
		transactionFactory: transactionFactory,
	}
}
