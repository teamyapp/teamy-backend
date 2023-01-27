package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type Thread struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.Thread = (*Thread)(nil)

func (t Thread) CreateThread(ct context.Context, threadID uint64) error {
	_, err := t.db.Exec(`
		INSERT INTO thread (id)
		VALUES ($1);
		`,
		threadID)

	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func (t Thread) DeleteThread(ct context.Context, threadID uint64) error {
	_, err := t.db.Exec(`
		DELETE FROM thread
		WHERE id = $1;
		`,
		threadID)

	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func NewThread(dataCollector telemetry.DataCollector, sqlDB *sql.DB) Thread {
	return Thread{dataCollector: dataCollector, db: sqlDB}
}
