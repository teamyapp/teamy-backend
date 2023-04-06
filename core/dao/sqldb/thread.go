package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type Thread struct {
	logger telemetry.Logger
	db     *sql.DB
}

var _ dao.Thread = (*Thread)(nil)

func (t Thread) CreateThread(ct context.Context, threadID uint64) *errs.Error {
	_, err := t.db.Exec(`
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

func (t Thread) DeleteThread(ct context.Context, threadID uint64) *errs.Error {
	_, err := t.db.Exec(`
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

func NewThread(logger telemetry.Logger, sqlDB *sql.DB) Thread {
	return Thread{logger: logger, db: sqlDB}
}
