package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
)

type RolloutVersionRelation interface {
	FindVersionNumbersByRolloutID(ct context.Context, rolloutID uint64) ([]int, *errs.Error)
}
