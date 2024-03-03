package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type PhaseStoryRelation struct {
	transactionFactory transaction.Factory
}

var _ dao.PhaseStoryRelation = (*PhaseStoryRelation)(nil)

func (p *PhaseStoryRelation) FindStoryIDsByPhaseIDWithTx(ct context.Context, tx *transaction.Transaction, phaseID uint64) ([]uint64, *errs.Error) {
	panic("unimplemented")
}

func (p *PhaseStoryRelation) CreatePhaseStoryRelation(ct context.Context, tx *transaction.Transaction, phaseStroyRelation entity.PhaseStoryRelation) *errs.Error {
	panic("unimplemented")
}

func (p *PhaseStoryRelation) DeletePhaseStoryRelation(ct context.Context, tx *transaction.Transaction, phaseID uint64, storyID uint64) *errs.Error {
	panic("unimplemented")
}

func NewPhaseStoryRelation(transactionFactory transaction.Factory) *PhaseStoryRelation {
	return &PhaseStoryRelation{
		transactionFactory: transactionFactory,
	}
}
