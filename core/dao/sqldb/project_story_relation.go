package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type ProjectStoryRelation struct {
	transactionFactory transaction.Factory
}

func (p *ProjectStoryRelation) FindStoryIDsByProjectIDWithTx(ct context.Context, tx *transaction.Transaction, projectID uint64) ([]uint64, *errs.Error) {
	panic("unimplemented")
}

func (p *ProjectStoryRelation) CreateProjectStoryRelation(ct context.Context, tx *transaction.Transaction, projectStoryRelation entity.ProjectStoryRelation) *errs.Error {
	panic("unimplemented")
}

func (p *ProjectStoryRelation) DeleteProjectStoryRelation(ct context.Context, tx *transaction.Transaction, projectID uint64, storyID uint64) *errs.Error {
	panic("unimplemented")
}

var _ dao.ProjectStoryRelation = (*ProjectStoryRelation)(nil)

func NewProjectStoryRelation(transactionFactory transaction.Factory) *ProjectStoryRelation {
	return &ProjectStoryRelation{
		transactionFactory: transactionFactory,
	}
}
