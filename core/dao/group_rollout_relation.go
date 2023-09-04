package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
)

type GroupRolloutRelation interface {
	FindRolloutIDsByGroupIDAndSortByOrderedIndex(ct context.Context, groupID uint64) ([]uint64, *errs.Error)
}
