package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type StoryLink interface {
	FindStoryLinkByID(ct context.Context, tx *transaction.Transaction, storyLinkID uint64) (entity.StoryLink, *errs.Error)
	FindLinksByStoryID(ct context.Context, tx *transaction.Transaction, storyID uint64) ([]entity.StoryLink, *errs.Error)
	CreateStoryLink(ct context.Context, tx *transaction.Transaction, storyLinkEntity entity.StoryLink) *errs.Error
	DeleteStoryLink(ct context.Context, tx *transaction.Transaction, storyLinkID uint64) *errs.Error
}
