package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type PercentageActivator struct{}

// CreatePercentageActivator implements dao.PercentageActivator.
func (*PercentageActivator) CreatePercentageActivator(ct context.Context, activator entity.PercentageActivator) (entity.PercentageActivator, *errs.Error) {
	panic("unimplemented")
}

// FindPercentageActivatorByID implements dao.PercentageActivator.
func (*PercentageActivator) FindPercentageActivatorByID(ct context.Context, ActivatorID uint64) (entity.PercentageActivator, *errs.Error) {
	panic("unimplemented")
}

var _ dao.PercentageActivator = (*PercentageActivator)(nil)

func NewPercentageActivator() *PercentageActivator {
	return &PercentageActivator{}
}
