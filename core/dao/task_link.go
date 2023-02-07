package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskLink interface {
	CreateTaskLink(ct context.Context, taskLinkEntity entity.TaskLink) *errs.Error
	FindLinksByTaskID(ct context.Context, taskID uint64) ([]entity.TaskLink, *errs.Error)
}
