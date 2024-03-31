package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type PhaseStoryRelation interface {
	FindStoryIDsByPhaseIDWithTx(ct context.Context, tx *transaction.Transaction, phaseID uint64) ([]uint64, *errs.Error)
	CreatePhaseStoryRelation(ct context.Context, tx *transaction.Transaction, phaseStroyRelation entity.PhaseStoryRelation) *errs.Error
	DeletePhaseStoryRelation(ct context.Context, tx *transaction.Transaction, phaseID uint64, storyID uint64) *errs.Error
	DeletePhaseStoryRelationsByPhaseID(ct context.Context, tx *transaction.Transaction, phaseID uint64) *errs.Error
	DeletePhaseStoryRelationsByStoryID(ct context.Context, tx *transaction.Transaction, storyID uint64) *errs.Error
}
