package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const appGroupRelationDaoName = "AppGroupRelation"

type AppGroupRelation struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.AppGroupRelation = (*AppGroupRelation)(nil)

func (a *AppGroupRelation) FindAppIDsByGroupID(ct context.Context, groupID uint64) ([]uint64, *errs.Error) {
	a.metrics.ReportDaoOperation(appGroupRelationDaoName, "FindAppIDsByGroupID")
	opt := sql.TxOptions{
		ReadOnly: true,
	}

	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return a.FindAppIDsByGroupIDWithTx(ct, tx, groupID)
}

func (a *AppGroupRelation) FindAppIDsByGroupIDWithTx(ct context.Context, tx *transaction.Transaction, groupID uint64) ([]uint64, *errs.Error) {
	a.metrics.ReportDaoOperation(appGroupRelationDaoName, "FindAppIDsByGroupIDWithTx")
	var appIDs []uint64
	rows, err := tx.SQLTx().QueryContext(ct,
		`
		SELECT
			app_id
		FROM app_group_relation
		WHERE group_id = $1`,
		groupID,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		var appID uint64
		err := rows.Scan(&appID)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		appIDs = append(appIDs, appID)
	}

	return appIDs, nil
}

func (a *AppGroupRelation) FindGroupIDsByAppIDWithTx(ct context.Context, tx *transaction.Transaction, appID uint64) ([]uint64, *errs.Error) {
	a.metrics.ReportDaoOperation(appGroupRelationDaoName, "FindGroupIDsByAppIDWithTx")
	var groupIDs []uint64
	rows, err := tx.SQLTx().QueryContext(ct,
		`
		SELECT
			group_id
		FROM app_group_relation
		WHERE app_id = $1`,
		appID,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
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

func (a *AppGroupRelation) FindGroupIDsByAppID(ct context.Context, appID uint64) ([]uint64, *errs.Error) {
	a.metrics.ReportDaoOperation(appGroupRelationDaoName, "FindGroupIDsByAppID")
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return a.FindGroupIDsByAppIDWithTx(ct, tx, appID)
}

func (a *AppGroupRelation) CreateAppGroupRelation(ct context.Context, tx *transaction.Transaction, appGroupRelation entity.AppGroupRelation) *errs.Error {
	a.metrics.ReportDaoOperation(appGroupRelationDaoName, "CreateAppGroupRelation")
	_, err := tx.SQLTx().ExecContext(ct,
		`INSERT INTO app_group_relation (
			app_id,
			group_id
		)
		VALUES (
			$1,
			$2
		)`,
		appGroupRelation.AppID,
		appGroupRelation.GroupID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a *AppGroupRelation) DeleteAppGroupRelation(ct context.Context, tx *transaction.Transaction, appID uint64, groupID uint64) *errs.Error {
	a.metrics.ReportDaoOperation(appGroupRelationDaoName, "DeleteAppGroupRelation")
	_, err := tx.SQLTx().ExecContext(ct,
		`
		DELETE FROM app_group_relation
		WHERE app_id = $1 AND group_id = $2`,
		appID,
		groupID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a *AppGroupRelation) DeleteAppGroupRelationsByGroupID(ct context.Context, tx *transaction.Transaction, groupID uint64) *errs.Error {
	a.metrics.ReportDaoOperation(appGroupRelationDaoName, "DeleteAppGroupRelationsByGroupID")
	_, err := tx.SQLTx().ExecContext(ct,
		`
		DELETE FROM app_group_relation
		WHERE group_id = $1`,
		groupID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewAppGroupRelation(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *AppGroupRelation {
	return &AppGroupRelation{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
