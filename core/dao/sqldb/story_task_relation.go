package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type StoryTaskRelation struct {
	transactionFactory transaction.Factory
}

var _ dao.StoryTaskRelation = (*StoryTaskRelation)(nil)

func (s *StoryTaskRelation) FindTaskIDsByStoryIDWithTx(ct context.Context, tx *transaction.Transaction, storyID uint64) ([]uint64, *errs.Error) {
	panic("unimplemented")
}

func (s *StoryTaskRelation) CreateStoryTaskRelation(ct context.Context, tx *transaction.Transaction, storyTaskRelation entity.StoryTaskRelation) *errs.Error {
	panic("unimplemented")
}

func (s *StoryTaskRelation) DeleteStoryTaskRelation(ct context.Context, tx *transaction.Transaction, storyID uint64, taskID uint64) *errs.Error {
	panic("unimplemented")
}

func NewStoryTaskRelation(transactionFactory transaction.Factory) *StoryTaskRelation {
	return &StoryTaskRelation{
		transactionFactory: transactionFactory,
	}
}
