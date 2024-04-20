package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
)

const threadDaoName = "Thread"

type Thread struct {
	metrics dao.Metrics
}

var _ dao.Thread = (*Thread)(nil)

func (t Thread) CreateThread(ct context.Context, tx *transaction.Transaction, threadID uint64) *errs.Error {
	t.metrics.ReportDaoOperation(threadDaoName, "CreateThread")
	_, err := tx.SQLTx().Exec(`
		INSERT INTO thread (id)
		VALUES ($1);
		`,
		threadID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t Thread) DeleteThread(ct context.Context, tx *transaction.Transaction, threadID uint64) *errs.Error {
	t.metrics.ReportDaoOperation(threadDaoName, "DeleteThread")
	_, err := tx.SQLTx().Exec(`
		DELETE FROM thread
		WHERE id = $1;
		`,
		threadID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewThread(metrics dao.Metrics) Thread {
	return Thread{
		metrics: metrics,
	}
}
