package dao

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubPullRequest interface {
	FindPullRequestByGithubNodeID(ct context.Context, githubNodeID string) (entity.GithubPullRequest, *errs.Error)
	FindAllPullRequests(ct context.Context) ([]entity.GithubPullRequest, *errs.Error)
	CreatePullRequest(ct context.Context, pullRequest entity.GithubPullRequest) *errs.Error
	UpdatePullRequest(ct context.Context, pullRequest entity.GithubPullRequest) *errs.Error
	DeletePullRequestByGithubNodeID(ct context.Context, githubNodeID string) *errs.Error
}
