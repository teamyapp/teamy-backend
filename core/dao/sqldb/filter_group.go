package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type FilterGroup struct {
}

var _ dao.FilterGroup = (*FilterGroup)(nil)

func (*FilterGroup) FindFilterGroupByID(ct context.Context, groupID uint64) (entity.FilterGroup, *errs.Error) {
	panic("unimplemented")
}

func (*FilterGroup) FindFilterGroupsByIDs(ct context.Context, groupID []uint64) ([]entity.FilterGroup, *errs.Error) {
	panic("unimplemented")
}

func (*FilterGroup) CreateFilterGroup(ct context.Context, group entity.FilterGroup) (entity.FilterGroup, *errs.Error) {
	panic("unimplemented")
}

func (*FilterGroup) UpdateFilterGroup(ct context.Context, group entity.FilterGroup) *errs.Error {
	panic("unimplemented")
}

func NewFilterGroup() *FilterGroup {
	return &FilterGroup{}
}
