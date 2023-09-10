package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type StaticGroup struct{}

// CreateStaticGroup implements dao.StaticGroup.
func (*StaticGroup) CreateStaticGroup(ct context.Context, group entity.StaticGroup) (entity.StaticGroup, *errs.Error) {
	panic("unimplemented")
}

// FindStaticGroupsByID implements dao.StaticGroup.
func (*StaticGroup) FindStaticGroupsByID(ct context.Context, groupID uint64) (entity.StaticGroup, *errs.Error) {
	panic("unimplemented")
}

// FindStaticGroupsByIDs implements dao.StaticGroup.
func (*StaticGroup) FindStaticGroupsByIDs(ct context.Context, groupIDs []uint64) ([]entity.StaticGroup, *errs.Error) {
	panic("unimplemented")
}

// UpdateStaticGroup implements dao.StaticGroup.
func (*StaticGroup) UpdateStaticGroup(ct context.Context, Group entity.StaticGroup) *errs.Error {
	panic("unimplemented")
}

var _ dao.StaticGroup = (*StaticGroup)(nil)

func NewStaticGroup() *StaticGroup {
	return &StaticGroup{}
}
