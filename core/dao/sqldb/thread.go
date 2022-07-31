package sqldb

import (
	"database/sql"

	"github.com/teamyapp/teamy-backend/core/dao"
)

type Thread struct {
	db *sql.DB
}

var _ dao.Thread = (*Thread)(nil)

func (t Thread) CreateThread(threadID uint64) error {
	_, err := t.db.Exec(`
		INSERT INTO thread (id)
		VALUES ($1);
		`,
		threadID)
	return err
}

func (t Thread) DeleteThread(threadID uint64) error {
	_, err := t.db.Exec(`
		DELETE FROM thread
		WHERE id = $1;
		`,
		threadID)
	return err
}

func NewThread(sqlDB *sql.DB) Thread {
	return Thread{db: sqlDB}
}
