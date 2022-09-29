package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubCodeReview interface {
	FindCodeReviewByGithubReviewerID(ct context.Context, pullRequestNodeID string, githubReviewerID uint64) (entity.GithubCodeReview, error)
	FindCodeReviewByInternalTaskID(ct context.Context, internalTaskID uint64) (entity.GithubCodeReview, error)
	CreateCodeReview(ct context.Context, codeReview entity.GithubCodeReview) error
	UpdateCodeReview(ct context.Context, codeReview entity.GithubCodeReview) error
	DeleteCodeReviewByInternalTaskID(ct context.Context, internalTaskID uint64) error
	DeleteCodeReviewByGithubReviewerID(ct context.Context, githubReviewerID uint64) error
}
