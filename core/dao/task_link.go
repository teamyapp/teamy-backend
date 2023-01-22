package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskLink interface {
	CreateTaskLink(ct context.Context, taskLinkEntity entity.TaskLink) error
	FindTaskLinksByTaskID(ct context.Context, taskID uint64) ([]entity.TaskLink, error)
}
