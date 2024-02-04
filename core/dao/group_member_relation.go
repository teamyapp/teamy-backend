package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type GroupMemberRelation interface {
	FindMemberIDsByGroupIDWithTx(ct context.Context, tx *transaction.Transaction, groupID uint64) ([]uint64, *errs.Error)
	FindMemberIDsByGroupID(ct context.Context, groupID uint64) ([]uint64, *errs.Error)
	FilterGroupIDsByMemberIDWithTx(ct context.Context, tx *transaction.Transaction, groupIDs []uint64, memberID uint64) ([]uint64, *errs.Error)
	FilterGroupIDsByMemberID(ct context.Context, groupIDs []uint64, memberID uint64) ([]uint64, *errs.Error)
	CreateGroupMemberRelation(ct context.Context, tx *transaction.Transaction, groupMemberRelation entity.GroupMemberRelation) *errs.Error
	DeleteGroupMemberRelation(ct context.Context, tx *transaction.Transaction, memberID uint64, groupID uint64) *errs.Error
	DeleteGroupMemberRelationsByGroupID(ct context.Context, tx *transaction.Transaction, groupID uint64) *errs.Error
}
