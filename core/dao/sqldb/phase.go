package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Phase struct {
	transactionFactory transaction.Factory
}

var _ dao.Phase = (*Phase)(nil)

func (p *Phase) FindPhaseByIDWithTx(ct context.Context, tx *transaction.Transaction, phaseID uint64) (entity.Phase, *errs.Error) {
	panic("unimplemented")
}

func (p *Phase) FindPhasesByIDsWithTx(ct context.Context, tx *transaction.Transaction, phaseIDs []uint64) ([]entity.Phase, *errs.Error) {
	panic("unimplemented")
}

func (p *Phase) CreatePhase(ct context.Context, tx *transaction.Transaction, phase entity.Phase) *errs.Error {
	panic("unimplemented")
}

func (p *Phase) UpdatePhase(ct context.Context, tx *transaction.Transaction, phase entity.Phase) *errs.Error {
	panic("unimplemented")
}

func (p *Phase) DeletePhase(ct context.Context, tx *transaction.Transaction, phaseID uint64) *errs.Error {
	panic("unimplemented")
}

func NewPhase(transactionFactory transaction.Factory) *Phase {
	return &Phase{
		transactionFactory: transactionFactory,
	}
}
