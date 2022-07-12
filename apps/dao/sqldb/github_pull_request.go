package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubPullRequest struct {
	db *sql.DB
}

var _ dao.GithubPullRequest = (*GithubPullRequest)(nil)

func (g GithubPullRequest) FindPullRequestByInternalTaskID(internalTaskID uint64) (entity.GithubPullRequest, error) {
	pullRequest := entity.GithubPullRequest{}
	err := g.db.QueryRow(`
	SELECT
	    internal_task_id,
	    github_pull_request_node_id
	FROM apps_github_pull_request
	WHERE internal_task_id = $1;
`,
		internalTaskID).
		Scan(
			&pullRequest.InternalTaskID,
			&pullRequest.NodeID,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.GithubPullRequest{}, dao.ErrNotFound(fmt.Sprintf(
			"GithubPullRequest not found: internalTaskID=%v", internalTaskID))
	}

	return pullRequest, err
}

func (g GithubPullRequest) FindPullRequestByGithubNodeID(githubNodeID string) (entity.GithubPullRequest, error) {
	pullRequest := entity.GithubPullRequest{}
	err := g.db.QueryRow(`
	SELECT
	    internal_task_id,
	    github_pull_request_node_id
	FROM apps_github_pull_request
	WHERE github_pull_request_node_id = $1;
`,
		githubNodeID).
		Scan(
			&pullRequest.InternalTaskID,
			&pullRequest.NodeID,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.GithubPullRequest{}, dao.ErrNotFound(fmt.Sprintf(
			"GithubPullRequest not found: githubNodeID=%v", githubNodeID))
	}

	return pullRequest, err
}

func (g GithubPullRequest) CreatePullRequest(pullRequest entity.GithubPullRequest) error {
	_, err := g.db.Exec(`
	INSERT INTO apps_github_pull_request
	(
	    internal_task_id,
	    github_pull_request_node_id
	)
	VALUES ($1, $2);
`,
		pullRequest.InternalTaskID,
		pullRequest.NodeID,
	)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (g GithubPullRequest) DeletePullRequestByInternalTaskID(internalTaskID uint64) error {
	_, err := g.db.Exec(`
		DELETE FROM apps_github_pull_request
		WHERE internal_task_id = $1;
		`,
		internalTaskID)
	return err
}

func (g GithubPullRequest) DeletePullRequestByGithubNodeID(githubNodeID string) error {
	_, err := g.db.Exec(`
		DELETE FROM apps_github_pull_request
		WHERE github_pull_request_node_id = $1;
		`,
		githubNodeID)
	return err
}

func NewGithubPullRequest(sqlDB *sql.DB) GithubPullRequest {
	return GithubPullRequest{db: sqlDB}
}
