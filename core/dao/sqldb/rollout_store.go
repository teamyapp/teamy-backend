package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type RolloutStore struct{}

// FindRolloutStoreByID implements dao.RolloutStore.
func (*RolloutStore) FindRolloutStoreByID(ct context.Context, storeID uint64) (entity.RolloutStore, *errs.Error) {
	panic("unimplemented")
}

// UpdateRolloutStore implements dao.RolloutStore.
func (*RolloutStore) UpdateRolloutStore(ct context.Context, store entity.RolloutStore) *errs.Error {
	panic("unimplemented")
}

var _ dao.RolloutStore = (*RolloutStore)(nil)

func NewRolloutStore() *RolloutStore {
	return &RolloutStore{}
}
