package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
)

type Thread interface {
	CreateThread(ct context.Context, threadID uint64) *errs.Error
	DeleteThread(ct context.Context, threadID uint64) *errs.Error
}
