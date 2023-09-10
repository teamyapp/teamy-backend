package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type ActivatorTypeRelation interface {
	FindActivatorTypeByID(ct context.Context, activatorID uint64) (entity.ActivatorType, *errs.Error)
}
