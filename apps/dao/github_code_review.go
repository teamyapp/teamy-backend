package dao

import (
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubCodeReview interface {
	FindCodeReviewByGithubReviewerID(pullRequestNodeID string, githubReviewerID uint64) (entity.GithubCodeReview, error)
	FindCodeReviewByInternalTaskID(internalTaskID uint64) (entity.GithubCodeReview, error)
	CreateCodeReview(codeReview entity.GithubCodeReview) error
	UpdateCodeReview(codeReview entity.GithubCodeReview) error
	DeleteCodeReviewByInternalTaskID(internalTaskID uint64) error
	DeleteCodeReviewByGithubReviewerID(githubReviewerID uint64) error
}
