package dao

import (
	"context"

	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubPullRequest interface {
	FindPullRequestByInternalTaskID(ct context.Context, internalTaskID uint64) (entity.GithubPullRequest, error)
	FindPullRequestByGithubNodeID(ct context.Context, githubNodeID string) (entity.GithubPullRequest, error)
	CreatePullRequest(ct context.Context, pullRequest entity.GithubPullRequest) error
	DeletePullRequestByInternalTaskID(ct context.Context, internalTaskID uint64) error
	DeletePullRequestByGithubNodeID(ct context.Context, githubNodeID string) error
}
