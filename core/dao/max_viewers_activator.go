package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type MaxViewersActivator interface {
	FindMaxViewersActivatorByID(ct context.Context, ActivatorID uint64) (entity.MaxViewersActivator, *errs.Error)
	CreateMaxViewersActivator(ct context.Context, activator entity.MaxViewersActivator) (entity.MaxViewersActivator, *errs.Error)
}
