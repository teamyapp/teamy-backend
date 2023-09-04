package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Group interface {
	FindGroupsByIDs(ct context.Context, GroupIDs []uint64) ([]entity.Group, *errs.Error)
	CreateGroup(ct context.Context, Group entity.Group) (entity.Group, *errs.Error)
	UpdateGroup(ct context.Context, GroupID uint64, Group entity.Group) *errs.Error
}
