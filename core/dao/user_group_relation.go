package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type UserGroupRelation interface {
	FindUserIDsByGroupID(ct context.Context, groupID uint64) ([]uint64, *errs.Error)
	CreateUserGroupRelation(ct context.Context, tx *transaction.Transaction, userGroupRelation entity.UserGroupRelation) *errs.Error
	DeleteUserGroupRelation(ct context.Context, tx *transaction.Transaction, groupID uint64, userID uint64) *errs.Error
}
