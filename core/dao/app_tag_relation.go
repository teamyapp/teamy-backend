package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppTagRelation interface {
	FindAppTagByAppIDAndTagIDRelationWithTx(
		ct context.Context,
		tx *transaction.Transaction,
		appID uint64,
		tagID uint64,
	) (entity.AppTagRelation, *errs.Error)
	FindTagIDsByAppIDWithTx(
		ct context.Context,
		tx *transaction.Transaction,
		appID uint64,
	) ([]uint64, *errs.Error)
	CreateAppTagRelation(
		ct context.Context,
		tx *transaction.Transaction,
		appTagRelation entity.AppTagRelation) *errs.Error
	DeleteAppTagRelationByAppIDAndTagID(
		ct context.Context,
		tx *transaction.Transaction,
		appID uint64,
		tagID uint64,
	) *errs.Error
}
