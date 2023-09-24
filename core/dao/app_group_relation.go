package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppGroupRelation interface {
	FindAppGroupRelationTypeWithTx(ct context.Context, tx *transaction.Transaction, appID uint64, groupID uint64) (entity.AppGroupRelationType, *errs.Error)
	FindAppGroupRelationType(ct context.Context, appID uint64, groupID uint64) (entity.AppGroupRelationType, *errs.Error)
	FindAppIDByGroupIDWithTx(ct context.Context, tx *transaction.Transaction, groupID uint64) (uint64, *errs.Error)
	FindAppIDByGroupID(ct context.Context, groupID uint64) (uint64, *errs.Error)
	FindGroupIDsByAppIDAndRelationType(ct context.Context, appID uint64, appGroupRelationType entity.AppGroupRelationType) ([]uint64, *errs.Error)
	CreateAppGroupRelation(ct context.Context, tx *transaction.Transaction, appGroupRelation entity.AppGroupRelation) *errs.Error
	DeleteAppGroupRelation(ct context.Context, tx *transaction.Transaction, appID uint64, groupID uint64) *errs.Error
}
