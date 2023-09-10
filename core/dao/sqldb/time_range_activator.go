package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TimeRangeActivator struct{}

// CreateTimeRangeActivator implements dao.TimeRangeActivator.
func (*TimeRangeActivator) CreateTimeRangeActivator(ct context.Context, activator entity.TimeRangeActivator) (entity.TimeRangeActivator, *errs.Error) {
	panic("unimplemented")
}

// FindTimeRangeActivatorByID implements dao.TimeRangeActivator.
func (*TimeRangeActivator) FindTimeRangeActivatorByID(ct context.Context, activatorID uint64) (entity.TimeRangeActivator, *errs.Error) {
	panic("unimplemented")
}

var _ dao.TimeRangeActivator = (*TimeRangeActivator)(nil)

func NewTimeRangeActivator() *TimeRangeActivator {
	return &TimeRangeActivator{}
}
