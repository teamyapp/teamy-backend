package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamGroupRelation interface {
	FindTeamIDsByGroupIDWithTx(ct context.Context, tx *transaction.Transaction, groupID uint64) ([]uint64, *errs.Error)
	FindTeamIDsByGroupID(ct context.Context, groupID uint64) ([]uint64, *errs.Error)
	CreateTeamGroupRelation(ct context.Context, tx *transaction.Transaction, teamGroupRelation entity.TeamGroupRelation) *errs.Error
	DeleteTeamGroupRelation(ct context.Context, tx *transaction.Transaction, teamID uint64, groupID uint64) *errs.Error
}
