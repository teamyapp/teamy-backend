package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const appSecretDaoName = "AppSecret"

type AppSecret struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

func (a *AppSecret) FindAppSecretByIDWithTx(ct context.Context, tx *transaction.Transaction, appSecretID uint64) (entity.AppSecret, *errs.Error) {
	a.metrics.ReportDaoOperation(appSecretDaoName, "FindAppSecretByIDWithTx")
	appSecret := entity.AppSecret{}
	err := tx.SQLTx().QueryRowContext(
		ct,
		`
		SELECT
			id,
			app_id,
			name,
			added_at,
			added_by_user_id,
			last_used_at
		FROM app_secret
		WHERE id = $1`,
		appSecretID,
	).Scan(
		&appSecret.ID,
		&appSecret.AppID,
		&appSecret.Name,
		&appSecret.AddedAt,
		&appSecret.AddedByUserID,
		&appSecret.LastUsedAt,
	)

	if err != nil {
		return entity.AppSecret{}, errs.NewError(errs.Unknown, err.Error())
	}

	return appSecret, nil
}

func (a *AppSecret) FindAppSecretsByAppIDWithTx(ct context.Context, tx *transaction.Transaction, appID uint64) ([]entity.AppSecret, *errs.Error) {
	a.metrics.ReportDaoOperation(appSecretDaoName, "FindAppSecretsByAppIDWithTx")
	rows, err := tx.SQLTx().QueryContext(ct,
		`
		SELECT
			id,
			app_id,
			name,
			added_at,
			added_by_user_id,
			last_used_at
		FROM app_secret
		WHERE app_id = $1`,
		appID,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	var appSecrets []entity.AppSecret
	for rows.Next() {
		var appSecret entity.AppSecret
		err := rows.Scan(
			&appSecret.ID,
			&appSecret.AppID,
			&appSecret.Name,
			&appSecret.AddedAt,
			&appSecret.AddedByUserID,
			&appSecret.LastUsedAt,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		appSecrets = append(appSecrets, appSecret)
	}

	return appSecrets, nil
}

func (a *AppSecret) FindSecretsByAppID(ct context.Context, appID uint64) ([]entity.AppSecret, *errs.Error) {
	a.metrics.ReportDaoOperation(appSecretDaoName, "FindSecretsByAppID")
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return a.FindAppSecretsByAppIDWithTx(ct, tx, appID)
}

func (a *AppSecret) CreateAppSecret(ct context.Context, tx *transaction.Transaction, appSecret entity.AppSecret) *errs.Error {
	a.metrics.ReportDaoOperation(appSecretDaoName, "CreateAppSecret")
	_, err := tx.SQLTx().ExecContext(ct,
		`
		INSERT INTO app_secret (
			id,
			app_id,
			name,
			secret,
			added_at,
			added_by_user_id,
			last_used_at
		)
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7
		)`,
		appSecret.ID,
		appSecret.AppID,
		appSecret.Name,
		appSecret.Secret,
		appSecret.AddedAt,
		appSecret.AddedByUserID,
		appSecret.LastUsedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a *AppSecret) UpdateAppSecret(ct context.Context, tx *transaction.Transaction, appSecretID uint64, appSecret entity.AppSecret) *errs.Error {
	a.metrics.ReportDaoOperation(appSecretDaoName, "UpdateAppSecret")
	_, err := tx.SQLTx().ExecContext(ct,
		`
		UPDATE app_secret SET
			name = $1,
			secret = $2,
			added_at = $3,
			added_by_user_id = $4,
			last_used_at = $5,
			app_id = $6
		WHERE id = $7`,
		appSecret.Name,
		appSecret.Secret,
		appSecret.AddedAt,
		appSecret.AddedByUserID,
		appSecret.LastUsedAt,
		appSecret.AppID,
		appSecretID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a *AppSecret) DeleteAppSecret(ct context.Context, tx *transaction.Transaction, appSecretID uint64) *errs.Error {
	a.metrics.ReportDaoOperation(appSecretDaoName, "DeleteAppSecret")
	_, err := tx.SQLTx().ExecContext(ct,
		`
		DELETE FROM app_secret
		WHERE id = $1`,
		appSecretID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a *AppSecret) DeleteAppSecretsByAppID(ct context.Context, tx *transaction.Transaction, appID uint64) *errs.Error {
	a.metrics.ReportDaoOperation(appSecretDaoName, "DeleteAppSecretsByAppID")
	_, err := tx.SQLTx().ExecContext(ct,
		`
		DELETE FROM app_secret
		WHERE app_id = $1`,
		appID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

var _ dao.AppSecret = (*AppSecret)(nil)

func NewAppSecret(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *AppSecret {
	return &AppSecret{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
