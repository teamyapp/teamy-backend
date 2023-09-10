package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type StaticGroup struct{}

var _ dao.StaticGroup = (*StaticGroup)(nil)

func (*StaticGroup) FindStaticGroupsByID(ct context.Context, groupID uint64) (entity.StaticGroup, *errs.Error) {
	panic("unimplemented")
}

func (*StaticGroup) FindStaticGroupsByIDs(ct context.Context, groupIDs []uint64) ([]entity.StaticGroup, *errs.Error) {
	panic("unimplemented")
}

func (*StaticGroup) CreateStaticGroup(ct context.Context, group entity.StaticGroup) (entity.StaticGroup, *errs.Error) {
	panic("unimplemented")
}

func (*StaticGroup) UpdateStaticGroup(ct context.Context, Group entity.StaticGroup) *errs.Error {
	panic("unimplemented")
}

func NewStaticGroup() *StaticGroup {
	return &StaticGroup{}
}
