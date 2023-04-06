package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
)

type Thread struct {
	logger telemetry.Logger
}

var _ daov2.Thread = (*Thread)(nil)

func (t Thread) CreateThread(ct context.Context, tx *transaction.Transaction, threadID uint64) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		INSERT INTO thread (id)
		VALUES ($1);
		`,
		threadID)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.logger.ErrorWithContext(ct, internalErr)
		return internalErr
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func NewThread(logger telemetry.Logger) Thread {
	return Thread{logger: logger}
}
