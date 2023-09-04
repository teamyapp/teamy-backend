package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type AppGroupRelation interface {
	FindAppIDByGroupID(ct context.Context, groupID uint64) (uint64, *errs.Error)
	FindGroupIDsByAppIDAndType(ct context.Context, appID uint64, appGroupRelationType entity.AppGroupRelationType) ([]uint64, *errs.Error)
	CreateAppGroupRelation(ct context.Context, appGroupRelation entity.AppGroupRelation) (entity.AppGroupRelation, *errs.Error)
}
