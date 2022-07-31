package dao

import (
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubPullRequest interface {
	FindPullRequestByInternalTaskID(internalTaskID uint64) (entity.GithubPullRequest, error)
	FindPullRequestByGithubNodeID(githubNodeID string) (entity.GithubPullRequest, error)
	CreatePullRequest(pullRequest entity.GithubPullRequest) error
	DeletePullRequestByInternalTaskID(internalTaskID uint64) error
	DeletePullRequestByGithubNodeID(githubNodeID string) error
}
