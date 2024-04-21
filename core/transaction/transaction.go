package transaction

import (
	"context"
	"database/sql"
	"time"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/realtime"
)

type GroupFactory struct {
	logger             telemetry.Logger
	metrics            Metrics
	transactionFactory transaction.Factory
	stateSyncer        *realtime.StateSyncer
}

// all unnecessary operations should be executed either before db txn begins or after db txn commits to avoid
// long-running txn
func (g GroupFactory) WithTransactionGroup(
	ct context.Context,
	readonly bool,
	execute func(
		tx *transaction.Transaction,
		rtTx *realtime.Transaction,
	) *errs.Error) *errs.Error {
	// If isolation level not specify explicitly here, it will be the default level of DB
	// For postgres, the default level is Read Committed
	// https://www.postgresql.org/docs/7.2/xact-read-committed.html

	start := time.Now()
	err := func() *errs.Error {
		opt := sql.TxOptions{
			ReadOnly: readonly,
		}
		tx, err := g.transactionFactory.BeginTx(ct, &opt)
		g.metrics.ReportTransactionBegin()
		defer tx.Rollback()
		if err != nil {
			return err
		}

		rtTx := realtime.NewTransaction(g.logger, g.stateSyncer)
		err = execute(tx, rtTx)
		if err != nil {
			return err
		}

		mutations := rtTx.GetMutations()
		for _, mutation := range mutations {
			internalErr := mutation.PrepareClientNotifiers(ct, tx)
			if internalErr != nil {
				return internalErr
			}
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		internalErr := rtTx.Notify(ct)
		if internalErr != nil {
			return internalErr
		}

		return nil
	}()
	duration := time.Since(start)
	if err == nil {
		g.metrics.ReportTransactionCommit()
		g.metrics.ReportTransactionDuration(duration, CommittedResult)
	} else {
		g.metrics.ReportTransactionRollback()
		g.metrics.ReportTransactionDuration(duration, RolledBackResult)
	}

	return err
}

func NewGroupFactory(
	logger telemetry.Logger,
	metrics Metrics,
	transactionFactory transaction.Factory,
	stateSyncer *realtime.StateSyncer,
) GroupFactory {
	return GroupFactory{
		logger:             logger,
		metrics:            metrics,
		transactionFactory: transactionFactory,
		stateSyncer:        stateSyncer,
	}
}
