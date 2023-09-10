package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type UserGroupRelation struct{}

var _ dao.UserGroupRelation = (*UserGroupRelation)(nil)

func (*UserGroupRelation) FindUserIDsByGroupID(ct context.Context, groupID uint64) ([]uint64, *errs.Error) {
	panic("unimplemented")
}
func (*UserGroupRelation) CreateUserGroupRelation(ct context.Context, userGroupRelation entity.UserGroupRelation) (entity.UserGroupRelation, *errs.Error) {
	panic("unimplemented")
}

func (*UserGroupRelation) DeleteUserGroupRelation(ct context.Context, groupID uint64, userID uint64) *errs.Error {
	panic("unimplemented")
}

func NewUserGroupRelation() *UserGroupRelation {
	return &UserGroupRelation{}
}
