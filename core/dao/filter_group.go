package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type FilterGroup interface {
	FindFilterGroupsByIDs(ct context.Context, groupID []uint64) ([]entity.FilterGroup, *errs.Error)
	FindFilterGroupByID(ct context.Context, groupID uint64) (entity.FilterGroup, *errs.Error)
	CreateFilterGroup(ct context.Context, tx *transaction.Transaction, group entity.FilterGroup) *errs.Error
	UpdateFilterGroup(ct context.Context, tx *transaction.Transaction, group entity.FilterGroup) *errs.Error
}
