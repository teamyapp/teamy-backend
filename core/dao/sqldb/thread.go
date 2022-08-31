package sqldb

import (
	"database/sql"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type Thread struct {
	dataCollector obs.DataCollector
	db            *sql.DB
}

var _ dao.Thread = (*Thread)(nil)

func (t Thread) CreateThread(threadID uint64) error {
	_, err := t.db.Exec(`
		INSERT INTO thread (id)
		VALUES ($1);
		`,
		threadID)

	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (t Thread) DeleteThread(threadID uint64) error {
	_, err := t.db.Exec(`
		DELETE FROM thread
		WHERE id = $1;
		`,
		threadID)

	if err != nil {
		t.dataCollector.Logger.Log(obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewThread(dataCollector obs.DataCollector, sqlDB *sql.DB) Thread {
	return Thread{dataCollector: dataCollector, db: sqlDB}
}
