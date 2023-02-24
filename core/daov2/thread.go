package daov2

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
)

type Thread interface {
	CreateThread(ct context.Context, tx *sql.Tx, threadID uint64) *errs.Error
	DeleteThread(ct context.Context, tx *sql.Tx, threadID uint64) *errs.Error
}
