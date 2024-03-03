package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Story struct {
	transactionFactory transaction.Factory
}

var _ dao.Story = (*Story)(nil)

func (s *Story) FindStoriesByIDsWithTx(ct context.Context, tx *transaction.Transaction, storyIDs []uint64) ([]entity.Story, *errs.Error) {
	panic("unimplemented")
}

func (s *Story) FindStoryByIDWithTx(ct context.Context, tx *transaction.Transaction, storyID uint64) (entity.Story, *errs.Error) {
	panic("unimplemented")
}

func (s *Story) CreateStory(ct context.Context, tx *transaction.Transaction, story entity.Story) *errs.Error {
	panic("unimplemented")
}

func (s *Story) UpdateStory(ct context.Context, tx *transaction.Transaction, story entity.Story) *errs.Error {
	panic("unimplemented")
}

func (s *Story) DeleteStory(ct context.Context, tx *transaction.Transaction, storyID uint64) *errs.Error {
	panic("unimplemented")
}

func NewStory(transactionFactory transaction.Factory) *Story {
	return &Story{
		transactionFactory: transactionFactory,
	}
}
