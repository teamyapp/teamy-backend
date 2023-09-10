package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type GroupRolloutRelation interface {
	FindRolloutIDsByGroupID(ct context.Context, groupID uint64) ([]uint64, *errs.Error)
	FindGroupRolloutRelationsByGroupID(ct context.Context, groupID uint64) ([]entity.GroupRolloutRelation, *errs.Error)
}
