package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Story interface {
	FindStoriesByIDsWithTx(ct context.Context, tx *transaction.Transaction, storyIDs []uint64) ([]entity.Story, *errs.Error)
	FindStoryByIDWithTx(ct context.Context, tx *transaction.Transaction, storyID uint64) (entity.Story, *errs.Error)
	CreateStory(ct context.Context, tx *transaction.Transaction, story entity.Story) *errs.Error
	UpdateStory(ct context.Context, tx *transaction.Transaction, story entity.Story) *errs.Error
	DeleteStory(ct context.Context, tx *transaction.Transaction, storyID uint64) *errs.Error
}
