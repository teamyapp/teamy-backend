package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Story struct {
	FindStoryByID             func(ct context.Context, storyID uint64) (entity.Story, *errs.Error)
	FindStoriesByTeamID       func(ct context.Context, teamID uint64) ([]entity.Story, *errs.Error)
	FindAllStories            func(ct context.Context) ([]entity.Story, *errs.Error)
	FindStoryByIDWithTx       func(ct context.Context, tx *transaction.Transaction, storyID uint64) (entity.Story, *errs.Error)
	FindStoriesByIDsWithTx    func(ct context.Context, tx *transaction.Transaction, storyIDs []uint64) ([]entity.Story, *errs.Error)
	FindStoriesByTeamIDWithTx func(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.Story, *errs.Error)
	FindAllStoriesWithTx      func(ct context.Context, tx *transaction.Transaction) ([]entity.Story, *errs.Error)
	CreateStory               func(ct context.Context, tx *transaction.Transaction, story entity.Story) *errs.Error
	UpdateStory               func(ct context.Context, tx *transaction.Transaction, story entity.Story) *errs.Error
	DeleteStory               func(ct context.Context, tx *transaction.Transaction, storyID uint64) *errs.Error
}
