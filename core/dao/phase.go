package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Phase interface {
	FindPhasesWithTx(ct context.Context, tx *transaction.Transaction) ([]entity.Phase, *errs.Error)
	FindPhasesByIDsWithTx(ct context.Context, tx *transaction.Transaction, phaseIDs []uint64) ([]entity.Phase, *errs.Error)
	FindPhaseByIDWithTx(ct context.Context, tx *transaction.Transaction, phaseID uint64) (entity.Phase, *errs.Error)
	CreatePhase(ct context.Context, tx *transaction.Transaction, phase entity.Phase) *errs.Error
	UpdatePhase(ct context.Context, tx *transaction.Transaction, phase entity.Phase) *errs.Error
	DeletePhase(ct context.Context, tx *transaction.Transaction, phaseID uint64) *errs.Error
	DeletePhasesByIDs(ct context.Context, tx *transaction.Transaction, phaseIDs []uint64) *errs.Error
}
