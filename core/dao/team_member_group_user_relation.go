package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	daoEntity "github.com/teamyapp/teamy-backend/core/dao/entity"
)

type TeamMemberGroupUserRelation interface {
	FindMemberGroupUserIDsByMemberGroupID(
		ct context.Context,
		tx *transaction.Transaction,
		memberGroupID uint64,
	) ([]uint64, *errs.Error)
	FindMemberGroupIDsByUserID(
		ct context.Context,
		tx *transaction.Transaction,
		userID uint64,
	) ([]uint64, *errs.Error)
	CreateMemberGroupUserRelation(
		ct context.Context,
		tx *transaction.Transaction,
		relation daoEntity.TeamMemberGroupUserRelation,
	) *errs.Error
	DeleteMemberGroupUserRelation(
		ct context.Context,
		tx *transaction.Transaction,
		relation daoEntity.TeamMemberGroupUserRelation,
	) *errs.Error
}
