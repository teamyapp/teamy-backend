package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Project struct {
	transactionFactory transaction.Factory
}

var _ dao.Project = (*Project)(nil)

func (p *Project) FindProjectByIDWithTx(ct context.Context, tx *transaction.Transaction, projectID uint64) (entity.Project, *errs.Error) {
	panic("implement me")
}

func (p *Project) CreateProject(ct context.Context, tx *transaction.Transaction, project entity.Project) *errs.Error {
	panic("implement me")
}

func (p *Project) UpdateProject(ct context.Context, tx *transaction.Transaction, project entity.Project) *errs.Error {
	panic("implement me")
}

func (p *Project) DeleteProject(ct context.Context, tx *transaction.Transaction, projectID uint64) *errs.Error {
	panic("implement me")
}

func NewProject(transactionFactory transaction.Factory) *Project {
	return &Project{
		transactionFactory: transactionFactory,
	}
}
