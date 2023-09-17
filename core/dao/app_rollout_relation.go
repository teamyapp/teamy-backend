package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppRolloutRelation interface {
	FindRolloutIDsByAppIDAndRelationType(ct context.Context, appID uint64, rolloutType entity.AppRolloutRelationType) ([]uint64, *errs.Error)
	CreateAppRolloutRelation(ct context.Context, tx *transaction.Transaction, appRolloutRelation entity.AppRolloutRelation) *errs.Error
}
