package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Group interface {
	FindGroupsByIDs(ct context.Context, groupIDs []uint64) ([]entity.Group, *errs.Error)
	FindGroupsByID(ct context.Context, groupID uint64) (entity.Group, *errs.Error)
	CreateGroup(ct context.Context, group entity.Group) (entity.Group, *errs.Error)
	UpdateGroup(ct context.Context, Group entity.Group) *errs.Error
}
