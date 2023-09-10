package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type MaxViewersActivator struct{}

var _ dao.MaxViewersActivator = (*MaxViewersActivator)(nil)

func (*MaxViewersActivator) FindMaxViewersActivatorByID(ct context.Context, ActivatorID uint64) (entity.MaxViewersActivator, *errs.Error) {
	panic("unimplemented")
}

func (*MaxViewersActivator) CreateMaxViewersActivator(ct context.Context, activator entity.MaxViewersActivator) (entity.MaxViewersActivator, *errs.Error) {
	panic("unimplemented")
}

func NewMaxViewersActivator() *MaxViewersActivator {
	return &MaxViewersActivator{}
}
