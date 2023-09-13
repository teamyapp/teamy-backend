package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Group struct{}

var _ dao.Group = (*Group)(nil)

func (*Group) FindGroupsByID(ct context.Context, groupID uint64) (entity.Group, *errs.Error) {
	panic("unimplemented")
}

func (*Group) FindGroupsByIDs(ct context.Context, groupIDs []uint64) ([]entity.Group, *errs.Error) {
	panic("unimplemented")
}

func (*Group) CreateGroup(ct context.Context, group entity.Group) (entity.Group, *errs.Error) {
	panic("unimplemented")
}

func (*Group) UpdateGroup(ct context.Context, Group entity.Group) *errs.Error {
	panic("unimplemented")
}

func NewGroup() *Group {
	return &Group{}
}
