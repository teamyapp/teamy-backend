package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type Thread struct {
}

var _ dao.Thread = (*Thread)(nil)

func (t Thread) CreateThread(ct context.Context, tx *transaction.Transaction, threadID uint64) *errs.Error {
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

func NewThread() Thread {
	return Thread{}
}
