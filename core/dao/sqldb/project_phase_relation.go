package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type ProjectPhaseRelation struct {
	transactionFactory transaction.Factory
}

func (p *ProjectPhaseRelation) FindPhaseIDsByProjectIDWithTx(ct context.Context, tx *transaction.Transaction, projectID uint64) ([]uint64, *errs.Error) {
	panic("unimplemented")
}

func (p *ProjectPhaseRelation) CreateProjectPhaseRelation(ct context.Context, tx *transaction.Transaction, projectPhaseRelation entity.ProjectPhaseRelation) *errs.Error {
	panic("unimplemented")
}

func (p *ProjectPhaseRelation) DeleteProjectPhaseRelation(ct context.Context, tx *transaction.Transaction, projectID uint64, phaseID uint64) *errs.Error {
	panic("unimplemented")
}

var _ dao.ProjectPhaseRelation = (*ProjectPhaseRelation)(nil)

func NewProjectPhaseRelation(transactionFactory transaction.Factory) *ProjectPhaseRelation {
	return &ProjectPhaseRelation{
		transactionFactory: transactionFactory,
	}
}
