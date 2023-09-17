package transaction

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type TransactionsContext struct {
	logger             telemetry.Logger
	transactionFactory transaction.Factory
	stateSyncer        *realtime.StateSyncer
	ct                 context.Context
}

// all unnecessary operations should be executed either before db txn begins or after db txn commits to avoid
// long-running txn
func (t TransactionsContext) WithTransactions(
	readonly bool,
	execute func(tx *transaction.Transaction, rtTx *realtime.Transaction,
	) *errs.Error) *errs.Error {
	// If isolation level not specify explicitly here, it will be the default level of DB
	// For postgres, the default level is Read Committed
	// https://www.postgresql.org/docs/7.2/xact-read-committed.html
	opt := sql.TxOptions{
		ReadOnly: readonly,
	}
	tx, err := t.transactionFactory.BeginTx(t.ct, &opt)
	defer tx.Rollback()
	if err != nil {
		return err
	}

	rtTx := realtime.NewTransaction(t.logger, t.stateSyncer)
	err = execute(tx, rtTx)
	if err != nil {
		return err
	}

	mutations := rtTx.GetMutations()
	for _, mutation := range mutations {
		internalErr := mutation.PrepareClientNotifiers(t.ct, tx)
		if internalErr != nil {
			return internalErr
		}
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	internalErr := rtTx.Notify(t.ct)
	if internalErr != nil {
		return internalErr
	}

	return nil
}

func NewTransactionsContext(
	logger telemetry.Logger,
	transactionFactory transaction.Factory,
	stateSyncer *realtime.StateSyncer,
	ct context.Context,
) TransactionsContext {
	return TransactionsContext{
		logger:             logger,
		transactionFactory: transactionFactory,
		stateSyncer:        stateSyncer,
		ct:                 ct,
	}
}
