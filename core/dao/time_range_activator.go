package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TimeRangeActivator interface {
	FindTimeRangeActivatorByID(ct context.Context, activatorID uint64) (entity.TimeRangeActivator, *errs.Error)
	CreateTimeRangeActivator(ct context.Context, activator entity.TimeRangeActivator) (entity.TimeRangeActivator, *errs.Error)
}
