package service

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/realtimev2"
)

type TransactionsContext struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
	stateSyncer   *realtimev2.StateSyncer
	ct            context.Context
}

// all unnecessary operations should be executed either before db txn begins or after db txn commits to avoid
// long-running txn
func (t TransactionsContext) withTransactions(execute func(sqlTx *sql.Tx, rtTx *realtimev2.Transaction) *errs.Error) *errs.Error {
	sqlTx, err := t.db.BeginTx(t.ct, nil)
	defer sqlTx.Rollback()
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.ErrorWithContext(t.ct, internalErr)
		return internalErr
	}

	rtTx := realtimev2.NewTransaction(t.dataCollector, t.stateSyncer)
	execute(sqlTx, rtTx)
	err = sqlTx.Commit()
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.ErrorWithContext(t.ct, internalErr)
		return internalErr
	}

	internalErr := rtTx.Notify(t.ct)
	if internalErr != nil {
		t.dataCollector.Logger.ErrorWithContext(t.ct, internalErr)
		return internalErr
	}

	return nil
}
