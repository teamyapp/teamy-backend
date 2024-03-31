package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type StoryTaskRelation interface {
	FindTaskIDsByStoryIDWithTx(ct context.Context, tx *transaction.Transaction, storyID uint64) ([]uint64, *errs.Error)
	CreateStoryTaskRelation(ct context.Context, tx *transaction.Transaction, storyTaskRelation entity.StoryTaskRelation) *errs.Error
	DeleteStoryTaskRelation(ct context.Context, tx *transaction.Transaction, storyID uint64, taskID uint64) *errs.Error
	DeleteStoryTaskRelationsByStoryID(ct context.Context, tx *transaction.Transaction, storyID uint64) *errs.Error
	DeleteStoryTaskRelationsByTaskID(ct context.Context, tx *transaction.Transaction, taskID uint64) *errs.Error
}
