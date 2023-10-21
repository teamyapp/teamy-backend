package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type StoryAwaitForRelation interface {
	FindAwaitingStoryIDsWithTx(ct context.Context, tx *transaction.Transaction, waitForStoryID uint64) ([]uint64, *errs.Error)
	FindAwaitForStoryIDsWithTx(ct context.Context, tx *transaction.Transaction, waitingStoryID uint64) ([]uint64, *errs.Error)
	CreateRelation(ct context.Context, tx *transaction.Transaction, relation entity.StoryAwaitForRelation) *errs.Error
	DeleteRelation(ct context.Context, tx *transaction.Transaction, waitingStoryID uint64, awaitForStoryID uint64) *errs.Error
}
