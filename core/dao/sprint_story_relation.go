package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type SprintStoryRelation interface {
	FindStoryIDsBySprintIDWithTx(ct context.Context, tx *transaction.Transaction, sprintID uint64) ([]uint64, *errs.Error)
	FindSprintIDsByStoryIDWithTx(ct context.Context, tx *transaction.Transaction, storyID uint64) ([]uint64, *errs.Error)
	CreateSprintStoryRelation(ct context.Context, tx *transaction.Transaction, relation entity.SprintStoryRelation) *errs.Error
	DeleteSprintStoryRelation(ct context.Context, tx *transaction.Transaction, sprintID uint64, storyID uint64) *errs.Error
}
