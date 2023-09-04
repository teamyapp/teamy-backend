package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppRolloutRelation interface {
	FindRolloutIDsByAppIDAndType(ct context.Context, appID uint64, rolloutType entity.AppRolloutRelationType) ([]uint64, *errs.Error)
}
