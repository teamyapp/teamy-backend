package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Activator interface {
	FindActivatorByID(ct context.Context, activatorID uint64) (entity.Activator, *errs.Error)
	CreateActivator(ct context.Context, activator entity.Activator) (entity.Activator, *errs.Error)
	UpdateActivator(ct context.Context, activator entity.Activator) *errs.Error
}
