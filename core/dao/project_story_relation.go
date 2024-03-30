package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type ProjectStoryRelation interface {
	FindStoryIDsByProjectIDWithTx(ct context.Context, tx *transaction.Transaction, projectID uint64) ([]uint64, *errs.Error)
	CreateProjectStoryRelation(ct context.Context, tx *transaction.Transaction, projectStoryRelation entity.ProjectStoryRelation) *errs.Error
	DeleteProjectStoryRelation(ct context.Context, tx *transaction.Transaction, projectID uint64, storyID uint64) *errs.Error
	DeleteProjectStoryRelationsByProjectID(ct context.Context, tx *transaction.Transaction, projectID uint64) *errs.Error
	DeleteProjectStoryRelationsByStoryID(ct context.Context, tx *transaction.Transaction, storyID uint64) *errs.Error
}
