package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type Thread struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
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
			Code:     dao.DBError,
			EmbedErr: err,
		}
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
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
			Code:     dao.DBError,
			EmbedErr: err,
		}
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func NewThread(dataCollector telemetry.DataCollector, sqlDB *sql.DB) Thread {
	return Thread{dataCollector: dataCollector, db: sqlDB}
}
