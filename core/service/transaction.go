package service

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type TransactionsContext struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
	stateSyncer   *realtime.StateSyncer
	ct            context.Context
}

// all unnecessary operations should be executed either before db txn begins or after db txn commits to avoid
// long-running txn
func (t TransactionsContext) withTransactions(readonly bool, execute func(sqlTx *sql.Tx, rtTx *realtime.Transaction) *errs.Error) *errs.Error {
	// If isolation level not specify explicitly here, it will be the default level of DB
	// For postgres, the default level is Read Committed
	// https://www.postgresql.org/docs/7.2/xact-read-committed.html
	opt := sql.TxOptions{
		ReadOnly: readonly,
	}
	sqlTx, err := t.db.BeginTx(t.ct, &opt)
	defer sqlTx.Rollback()
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.ErrorWithContext(t.ct, internalErr)
		return internalErr
	}

	rtTx := realtime.NewTransaction(t.dataCollector, t.stateSyncer)
	execute(sqlTx, rtTx)

	mutations := rtTx.GetMutations()
	for _, mutation := range mutations {
		internalErr := mutation.PrepareClientNotifiers(t.ct, sqlTx)
		if internalErr != nil {
			t.dataCollector.Logger.ErrorWithContext(t.ct, internalErr)
			return internalErr
		}
	}

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
