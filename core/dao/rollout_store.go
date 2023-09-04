package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type RolloutStore interface {
	FindRolloutStoreByID(ct context.Context, storeID uint64) (entity.RolloutStore, *errs.Error)
	UpdateRolloutStore(ct context.Context, store entity.RolloutStore) *errs.Error
}
