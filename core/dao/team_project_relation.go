package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamProjectRelation interface {
	FindProjectIDsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]uint64, *errs.Error)
	FindTeamIDsByProjectIDWithTx(ct context.Context, tx *transaction.Transaction, projectID uint64) ([]uint64, *errs.Error)
	CreateTeamProjectRelation(ct context.Context, tx *transaction.Transaction, teamProjectRelation entity.TeamProjectRelation) *errs.Error
	DeleteTeamProjectRelation(ct context.Context, tx *transaction.Transaction, teamID uint64, projectID uint64) *errs.Error
}
