package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Tag interface {
	FindTagsByTagIDsWithTx(ct context.Context, tx *transaction.Transaction, tagIDs []uint64) ([]entity.Tag, *errs.Error)
	FindTagByNameWithTx(ct context.Context, tx *transaction.Transaction, name string) (entity.Tag, *errs.Error)
	FindTagByIDWithTx(ct context.Context, tx *transaction.Transaction, tagID uint64) (entity.Tag, *errs.Error)
	CreateTag(ct context.Context, tx *transaction.Transaction, tag entity.Tag) *errs.Error
}
