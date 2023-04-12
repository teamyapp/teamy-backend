package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type Thread struct {
	db *sql.DB
}

var _ dao.Thread = (*Thread)(nil)

func (t Thread) CreateThread(ct context.Context, threadID uint64) *errs.Error {
	_, err := t.db.Exec(`
		INSERT INTO thread (id)
		VALUES ($1);
		`,
		threadID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
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
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewThread(sqlDB *sql.DB) Thread {
	return Thread{db: sqlDB}
}
