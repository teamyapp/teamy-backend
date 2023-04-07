package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubCodeReview interface {
	FindCodeReviewByGithubReviewerID(ct context.Context, pullRequestNodeID string, githubReviewerNodeID string) (entity.GithubCodeReview, *errs.Error)
	FindCodeReviewByInternalTaskID(ct context.Context, internalTaskID uint64) (entity.GithubCodeReview, *errs.Error)
	CreateCodeReview(ct context.Context, codeReview entity.GithubCodeReview) *errs.Error
	UpdateCodeReview(ct context.Context, codeReview entity.GithubCodeReview) *errs.Error
	DeleteCodeReviewByInternalTaskID(ct context.Context, internalTaskID uint64) *errs.Error
	DeleteCodeReviewByGithubReviewerID(ct context.Context, githubReviewerID uint64) *errs.Error
}
