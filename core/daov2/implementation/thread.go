package implementation

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/daov2"
)

type Thread struct {
	dataCollector telemetry.DataCollector
}

var _ daov2.Thread = (*Thread)(nil)

func (t Thread) CreateThread(ct context.Context, tx *sql.Tx, threadID uint64) *errs.Error {
	_, err := tx.Exec(`
		INSERT INTO thread (id)
		VALUES ($1);
		`,
		threadID)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (t Thread) DeleteThread(ct context.Context, tx *sql.Tx, threadID uint64) *errs.Error {
	_, err := tx.Exec(`
		DELETE FROM thread
		WHERE id = $1;
		`,
		threadID)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func NewThread(dataCollector telemetry.DataCollector) Thread {
	return Thread{dataCollector: dataCollector}
}
