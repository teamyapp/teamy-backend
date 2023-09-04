package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type FilterGroup interface {
	FindFilterGroupsByIDs(ct context.Context, groupID []uint64) ([]entity.FilterGroup, *errs.Error)
	CreateFilterGroup(ct context.Context, group entity.FilterGroup) (entity.FilterGroup, *errs.Error)
	UpdateFilterGroup(ct context.Context, groupID uint64, group entity.FilterGroup) *errs.Error
}
