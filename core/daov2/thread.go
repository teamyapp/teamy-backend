package daov2

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
)

type Thread interface {
	CreateThread(ct context.Context, tx *transaction.Transaction, threadID uint64) *errs.Error
	DeleteThread(ct context.Context, tx *transaction.Transaction, threadID uint64) *errs.Error
}
