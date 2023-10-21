package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type StoryTaskRelation interface {
	FindTaskIDsByStoryID(ct context.Context, storyID uint64) ([]uint64, *errs.Error)
	FindStoryIDsByTaskID(ct context.Context, taskID uint64) ([]uint64, *errs.Error)
	CreateStoryTaskRelation(ct context.Context, storyID uint64, taskID uint64, relation entity.StoryTaskRelation) *errs.Error
	DeleteStoryTaskRelation(ct context.Context, storyID uint64, taskID uint64) *errs.Error
}
