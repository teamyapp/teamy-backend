package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Group interface {
	FindGroupsByIDs(ct context.Context, groupIDs []uint64) ([]entity.Group, *errs.Error)
	FindGroupsByIDsWithTx(ct context.Context, tx *transaction.Transaction, groupIDs []uint64) ([]entity.Group, *errs.Error)
	FindGroupByID(ct context.Context, groupID uint64) (entity.Group, *errs.Error)
	FindGroupByIDWithTx(ct context.Context, tx *transaction.Transaction, groupID uint64) (entity.Group, *errs.Error)
	CreateGroup(ct context.Context, tx *transaction.Transaction, group entity.Group) *errs.Error
	UpdateGroup(ct context.Context, tx *transaction.Transaction, group entity.Group) *errs.Error
}
