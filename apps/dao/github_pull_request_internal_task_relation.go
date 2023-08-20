package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubPullRequestInternalTaskRelation interface {
	FindPullRequestInternalTaskRelationsByInternalTaskID(ct context.Context, internalTaskID uint64) ([]entity.GithubPullRequestInternalTaskRelation, *errs.Error)
	FindPullRequestInternalTaskRelationsByNodeID(ct context.Context, nodeID string) ([]entity.GithubPullRequestInternalTaskRelation, *errs.Error)
	CreatePullRequestInternalTaskRelation(ct context.Context, pullRequestInternalTaskRelation entity.GithubPullRequestInternalTaskRelation) *errs.Error
	DeletePullRequestInternalTaskRelationByNodeIDAndTaskID(ct context.Context, nodeID string, internalTaskID uint64) *errs.Error
}
