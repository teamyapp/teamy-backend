package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubPullRequest struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.GithubPullRequest = (*GithubPullRequest)(nil)

func (g GithubPullRequest) FindPullRequestByInternalTaskID(ct context.Context, internalTaskID uint64) (entity.GithubPullRequest, error) {
	pullRequest := entity.GithubPullRequest{}
	err := g.db.QueryRow(`
	SELECT
	    internal_task_id,
	    github_pull_request_node_id,
	    github_repository_owner
	    github_repository_name,
	    github_pull_request_number,
	    github_pull_request_url,
	FROM apps_github_pull_request
	WHERE internal_task_id = $1;
`,
		internalTaskID).
		Scan(
			&pullRequest.InternalTaskID,
			&pullRequest.NodeID,
			&pullRequest.RepositoryOwner,
			&pullRequest.RepositoryName,
			&pullRequest.Number,
			&pullRequest.URL,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.GithubPullRequest{}, dao.ErrNotFound(fmt.Sprintf(
			"GithubPullRequest not found: internalTaskID=%v", internalTaskID))
	}

	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return pullRequest, err
}

func (g GithubPullRequest) FindPullRequestByGithubNodeID(ct context.Context, githubNodeID string) (entity.GithubPullRequest, error) {
	pullRequest := entity.GithubPullRequest{}
	err := g.db.QueryRow(`
	SELECT
	    internal_task_id,
	    github_pull_request_node_id,
	    github_repository_owner,
	    github_repository_name,
	    github_pull_request_number,
	    github_pull_request_url,
	FROM apps_github_pull_request
	WHERE github_pull_request_node_id = $1;
`,
		githubNodeID).
		Scan(
			&pullRequest.InternalTaskID,
			&pullRequest.NodeID,
			&pullRequest.Number,
			&pullRequest.URL,
			&pullRequest.RepositoryName,
			&pullRequest.RepositoryOwner,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.GithubPullRequest{}, dao.ErrNotFound(fmt.Sprintf(
			"GithubPullRequest not found: githubNodeID=%v", githubNodeID))
	}

	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return pullRequest, err
}

func (g GithubPullRequest) CreatePullRequest(ct context.Context, pullRequest entity.GithubPullRequest) error {
	_, err := g.db.Exec(`
	INSERT INTO apps_github_pull_request
	(
	    internal_task_id,
	    github_pull_request_node_id,
	 	github_pull_request_number,
	    github_pull_request_url,
	    github_repository_name,
	    github_repository_owner
	)
	VALUES ($1, $2, $3, $4, $5, $6);
`,
		pullRequest.InternalTaskID,
		pullRequest.NodeID,
		pullRequest.Number,
		pullRequest.URL,
		pullRequest.RepositoryName,
		pullRequest.RepositoryOwner,
	)

	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func (g GithubPullRequest) DeletePullRequestByInternalTaskID(ct context.Context, internalTaskID uint64) error {
	_, err := g.db.Exec(`
		DELETE FROM apps_github_pull_request
		WHERE internal_task_id = $1;
		`,
		internalTaskID)

	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func (g GithubPullRequest) DeletePullRequestByGithubNodeID(ct context.Context, githubNodeID string) error {
	_, err := g.db.Exec(`
		DELETE FROM apps_github_pull_request
		WHERE github_pull_request_node_id = $1;
		`,
		githubNodeID)

	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func NewGithubPullRequest(dataCollector telemetry.DataCollector, sqlDB *sql.DB) GithubPullRequest {
	return GithubPullRequest{dataCollector: dataCollector, db: sqlDB}
}
