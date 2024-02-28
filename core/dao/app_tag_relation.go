package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppTagRelation interface {
	FindAppIDsByTagValuesWithTx(
		ct context.Context,
		tx *transaction.Transaction,
		tagValues []string,
	) ([]uint64, *errs.Error)
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
