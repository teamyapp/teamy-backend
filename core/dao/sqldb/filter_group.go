package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type FilterGroup struct {
}

// CreateFilterGroup implements dao.FilterGroup.
func (*FilterGroup) CreateFilterGroup(ct context.Context, group entity.FilterGroup) (entity.FilterGroup, *errs.Error) {
	panic("unimplemented")
}

// FindFilterGroupByID implements dao.FilterGroup.
func (*FilterGroup) FindFilterGroupByID(ct context.Context, groupID uint64) (entity.FilterGroup, *errs.Error) {
	panic("unimplemented")
}

// FindFilterGroupsByIDs implements dao.FilterGroup.
func (*FilterGroup) FindFilterGroupsByIDs(ct context.Context, groupID []uint64) ([]entity.FilterGroup, *errs.Error) {
	panic("unimplemented")
}

// UpdateFilterGroup implements dao.FilterGroup.
func (*FilterGroup) UpdateFilterGroup(ct context.Context, group entity.FilterGroup) *errs.Error {
	panic("unimplemented")
}

var _ dao.FilterGroup = (*FilterGroup)(nil)

func NewFilterGroup() *FilterGroup {
	return &FilterGroup{}
}
