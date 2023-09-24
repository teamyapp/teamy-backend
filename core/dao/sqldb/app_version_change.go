package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
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
		FROM app_version_changes 
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

func NewAppVersionChange() *AppVersionChange {
	return &AppVersionChange{}
}
