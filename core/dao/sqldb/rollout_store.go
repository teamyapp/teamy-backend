package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type RolloutStore struct{}

var _ dao.RolloutStore = (*RolloutStore)(nil)

func (*RolloutStore) FindRolloutStoreByID(ct context.Context, storeID uint64) (entity.RolloutStore, *errs.Error) {
	panic("unimplemented")
}

func (*RolloutStore) UpdateRolloutStore(ct context.Context, store entity.RolloutStore) *errs.Error {
	panic("unimplemented")
}

func NewRolloutStore() *RolloutStore {
	return &RolloutStore{}
}
