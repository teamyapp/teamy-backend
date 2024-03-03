package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type ProjectPhaseRelation interface {
	FindPhaseIDsByProjectIDWithTx(ct context.Context, tx *transaction.Transaction, projectID uint64) ([]uint64, *errs.Error)
	CreateProjectPhaseRelation(ct context.Context, tx *transaction.Transaction, projectPhaseRelation entity.ProjectPhaseRelation) *errs.Error
	DeleteProjectPhaseRelation(ct context.Context, tx *transaction.Transaction, projectID uint64, phaseID uint64) *errs.Error
}
