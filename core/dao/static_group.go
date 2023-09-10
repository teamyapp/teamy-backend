package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type StaticGroup interface {
	FindStaticGroupsByIDs(ct context.Context, groupIDs []uint64) ([]entity.StaticGroup, *errs.Error)
	FindStaticGroupsByID(ct context.Context, groupID uint64) (entity.StaticGroup, *errs.Error)
	CreateStaticGroup(ct context.Context, group entity.StaticGroup) (entity.StaticGroup, *errs.Error)
	UpdateStaticGroup(ct context.Context, Group entity.StaticGroup) *errs.Error
}
