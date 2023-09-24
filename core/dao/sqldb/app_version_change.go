package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppVersionChange struct {
	transactionFactory transaction.Factory
}

var _ dao.AppVersionChange = (*AppVersionChange)(nil)

func (a *AppVersionChange) FindAppVersionChangesByAppIDAndVersionNumberWithTx(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int) ([]string, *errs.Error) {
	rows, err := tx.SQLTx().QueryContext(
		ct,
		`
		SELECT change 
		FROM app_version_change 
		WHERE app_id = $1 AND version_number = $2",
		appID,
		`,
		versionNumber,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	var changes []string
	for rows.Next() {
		var change string
		err := rows.Scan(&change)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		changes = append(changes, change)
	}

	return changes, nil
}

func (a *AppVersionChange) FindAppVersionChangesByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int) ([]string, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}

	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return NewAppVersionChange().FindAppVersionChangesByAppIDAndVersionNumberWithTx(ct, tx, appID, versionNumber)
}

func (*AppVersionChange) CreateAppVersionChange(ct context.Context, tx *transaction.Transaction, appVersionChange entity.AppVersionChange) *errs.Error {
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
		INSERT INTO app_version_change (app_id, version_number, change) 
		VALUES ($1, $2, $3)
		`,
		appVersionChange.AppID,
		appVersionChange.VersionNumber,
		appVersionChange.Change,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewAppVersionChange() *AppVersionChange {
	return &AppVersionChange{}
}
