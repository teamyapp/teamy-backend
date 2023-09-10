package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TimeRangeActivator struct{}

var _ dao.TimeRangeActivator = (*TimeRangeActivator)(nil)

func (*TimeRangeActivator) FindTimeRangeActivatorByID(ct context.Context, activatorID uint64) (entity.TimeRangeActivator, *errs.Error) {
	panic("unimplemented")
}
func (*TimeRangeActivator) CreateTimeRangeActivator(ct context.Context, activator entity.TimeRangeActivator) (entity.TimeRangeActivator, *errs.Error) {
	panic("unimplemented")
}

func NewTimeRangeActivator() *TimeRangeActivator {
	return &TimeRangeActivator{}
}
