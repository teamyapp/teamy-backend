package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type GroupRolloutRelation interface {
	FindGroupRolloutRelationsByGroupIDWithTx(ct context.Context, tx *transaction.Transaction, groupID uint64) ([]entity.GroupRolloutRelation, *errs.Error)
	FindGroupRolloutRelationsByRolloutIDWithTx(ct context.Context, tx *transaction.Transaction, rolloutID uint64) ([]entity.GroupRolloutRelation, *errs.Error)
	FindGroupRolloutByGroupIDAndRolloutIDWithTx(ct context.Context, tx *transaction.Transaction, groupID, rolloutID uint64) (*entity.GroupRolloutRelation, *errs.Error)
	FindGroupRolloutRelationsByGroupID(ct context.Context, groupID uint64) ([]entity.GroupRolloutRelation, *errs.Error)
	FindGroupRolloutRelationsByRolloutID(ct context.Context, rolloutID uint64) ([]entity.GroupRolloutRelation, *errs.Error)
	CreateGroupRolloutRelation(ct context.Context, tx *transaction.Transaction, groupRolloutRelation entity.GroupRolloutRelation) *errs.Error
	DeleteGroupRolloutRelationsByGroupID(ct context.Context, tx *transaction.Transaction, groupID uint64) *errs.Error
	DeleteGroupRolloutRelationsByGroupIDAndRolloutID(ct context.Context, tx *transaction.Transaction, groupID, rolloutID uint64) *errs.Error
	UpdateFromOrderIndexByGroupID(ct context.Context, tx *transaction.Transaction, step int, orderIndex int, groupID uint64) *errs.Error
}
