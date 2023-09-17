package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type GroupRolloutRelation interface {
	FindGroupRolloutRelationsByGroupIDWithTx(ct context.Context, tx *transaction.Transaction, groupID uint64) ([]entity.GroupRolloutRelation, *errs.Error)
	FindGroupRolloutRelationsByGroupID(ct context.Context, groupID uint64) ([]entity.GroupRolloutRelation, *errs.Error)
	CreateGroupRolloutRelation(ct context.Context, tx *transaction.Transaction, groupRolloutRelation entity.GroupRolloutRelation) *errs.Error
}
