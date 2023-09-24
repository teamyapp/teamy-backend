package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppSecret struct {
	transactionFactory transaction.Factory
}

func (*AppSecret) FindAppSecretByIDWithTx(ct context.Context, tx *transaction.Transaction, appSecretID uint64) (entity.AppSecret, *errs.Error) {
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
		FROM app_secrets
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
	rows, err := tx.SQLTx().QueryContext(ct,
		`
		SELECT
			id,
			app_id,
			name,
			added_at,
			added_by_user_id,
			last_used_at
		FROM app_secrets
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

func (*AppSecret) CreateAppSecret(ct context.Context, tx *transaction.Transaction, appSecret entity.AppSecret) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`
		INSERT INTO app_secrets (
			app_id,
			name,
			added_at,
			added_by_user_id,
			last_used_at
		) 
		VALUES (
			$1,
			$2,
			$3,
			$4,
			$5
		)`,
		appSecret.AppID,
		appSecret.Name,
		appSecret.AddedAt,
		appSecret.AddedByUserID,
		appSecret.LastUsedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (*AppSecret) UpdateAppSecret(ct context.Context, tx *transaction.Transaction, appSecretID uint64, appSecret entity.AppSecret) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`
		UPDATE app_secrets SET
			name = $1,
			added_at = $2,
			added_by_user_id = $3,
			last_used_at = $4,
			app_id = $5
		WHERE id = $6`,
		appSecret.Name,
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

func (*AppSecret) DeleteAppSecret(ct context.Context, tx *transaction.Transaction, appSecretID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct,
		`
		DELETE FROM app_secrets
		WHERE id = $1`,
		appSecretID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

var _ dao.AppSecret = (*AppSecret)(nil)

func NewAppSecret(transactionFactory transaction.Factory) *AppSecret {
	return &AppSecret{
		transactionFactory: transactionFactory,
	}
}
