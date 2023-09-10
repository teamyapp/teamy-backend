package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppRolloutRelation interface {
	FindRolloutIDsByAppIDAndRelationType(ct context.Context, appID uint64, rolloutType entity.AppRolloutRelationType) ([]uint64, *errs.Error)
	CreateAppRolloutRelation(ct context.Context, appRolloutRelation entity.AppRolloutRelation) (entity.AppRolloutRelation, *errs.Error)
}
