package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppVersionPrice struct {
	transactionFactory *transaction.Factory
}

var _ dao.AppVersionPrice = (*AppVersionPrice)(nil)

func (a *AppVersionPrice) FindAppVersionPricesByAppIDAndVersionNumberWithTx(ct context.Context, tx *transaction.Transaction, appID uint64, versionNumber int) ([]entity.Money, *errs.Error) {
	rows, err := tx.SQLTx().QueryContext(ct,
		`
		SELECT currency, amount 
		FROM app_version_prices 
		WHERE app_id = $1 AND version_number = $2;",
		appID,
		versionNumber,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	var prices []entity.Money
	for rows.Next() {
		var price entity.Money
		err := rows.Scan(&price.Currency, &price.Amount)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		prices = append(prices, price)
	}

	return prices, nil
}

func (a *AppVersionPrice) FindAppVersionPricesByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int) ([]entity.Money, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}

	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return a.FindAppVersionPricesByAppIDAndVersionNumberWithTx(ct, tx, appID, versionNumber)
}

func NewAppVersionPrice(transactionFactory transaction.Factory) *AppVersionPrice {
	return &AppVersionPrice{
		transactionFactory: &transactionFactory,
	}
}
