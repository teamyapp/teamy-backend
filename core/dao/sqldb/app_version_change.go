package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const appVersionChangeDaoName = "AppVersionChange"

type AppVersionChange struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.AppVersionChange = (*AppVersionChange)(nil)

func (a *AppVersionChange) FindAppVersionChangesByAppIDAndVersionNumberWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	appID uint64,
	versionNumber int,
) ([]entity.AppVersionChange, *errs.Error) {
	a.metrics.ReportDaoOperation(appVersionChangeDaoName, "FindAppVersionChangesByAppIDAndVersionNumberWithTx")
	rows, err := tx.SQLTx().QueryContext(
		ct,
		`
		SELECT
			app_id,
			version_number,
			change
		FROM app_version_change
		WHERE app_id = $1 AND version_number = $2`,
		appID,
		versionNumber,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	var changes []entity.AppVersionChange
	for rows.Next() {
		change := entity.AppVersionChange{}
		err := rows.Scan(
			&change.AppID,
			&change.VersionNumber,
			&change.Change,
		)

		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		changes = append(changes, change)
	}

	return changes, nil
}

func (a *AppVersionChange) FindAppVersionChangesByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int) ([]entity.AppVersionChange, *errs.Error) {
	a.metrics.ReportDaoOperation(appVersionChangeDaoName, "FindAppVersionChangesByAppIDAndVersionNumber")
	opt := sql.TxOptions{
		ReadOnly: true,
	}

	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return a.FindAppVersionChangesByAppIDAndVersionNumberWithTx(ct, tx, appID, versionNumber)
}

func (a *AppVersionChange) CreateAppVersionChange(ct context.Context, tx *transaction.Transaction, appVersionChange entity.AppVersionChange) *errs.Error {
	a.metrics.ReportDaoOperation(appVersionChangeDaoName, "CreateAppVersionChange")
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
		INSERT INTO app_version_change (
		  id,
		  app_id,
		  version_number,
		  change
		)
		VALUES ($1, $2, $3, $4)
		`,
		appVersionChange.ID,
		appVersionChange.AppID,
		appVersionChange.VersionNumber,
		appVersionChange.Change,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewAppVersionChange(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *AppVersionChange {
	return &AppVersionChange{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
