package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubPullRequest struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.GithubPullRequest = (*GithubPullRequest)(nil)

func (g GithubPullRequest) FindPullRequestByGithubNodeID(
	ct context.Context,
	githubNodeID string,
) (entity.GithubPullRequest, *errs.Error) {
	pullRequest := entity.GithubPullRequest{}
	err := g.db.QueryRow(`
	SELECT
	    github_pull_request_node_id,
	    github_repository_owner,
	    github_repository_name,
	    github_pull_request_number,
	    github_pull_request_url,
	    github_organization_id
	FROM apps_github_pull_request
	WHERE github_pull_request_node_id = $1;
`,
		githubNodeID).
		Scan(
			&pullRequest.NodeID,
			&pullRequest.RepositoryOwner,
			&pullRequest.RepositoryName,
			&pullRequest.Number,
			&pullRequest.URL,
			&pullRequest.OrganizationID,
		)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.GithubPullRequest{}, errs.NewError(errs.NotFound, fmt.Sprintf(
				"GithubPullRequest not found: githubNodeID=%v",
				githubNodeID))
		}

		return entity.GithubPullRequest{}, errs.NewError(errs.Unknown, err.Error())
	}

	return pullRequest, nil
}

func (g GithubPullRequest) CreatePullRequest(
	ct context.Context,
	pullRequest entity.GithubPullRequest,
) *errs.Error {
	_, err := g.db.Exec(`
	INSERT INTO apps_github_pull_request
	(
	    github_pull_request_node_id,
	 	github_repository_owner,
	    github_repository_name,
	    github_pull_request_number,
	    github_pull_request_url,
	 	github_organization_id
	)
	VALUES ($1, $2, $3, $4, $5, $6);
`,
		pullRequest.NodeID,
		pullRequest.RepositoryOwner,
		pullRequest.RepositoryName,
		pullRequest.Number,
		pullRequest.URL,
		pullRequest.OrganizationID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (g GithubPullRequest) UpdatePullRequest(ct context.Context, pullRequest entity.GithubPullRequest) *errs.Error {
	_, err := g.db.Exec(`
		UPDATE apps_github_pull_request
		SET
    		github_repository_owner = $1,
			github_repository_name = $2,
			github_pull_request_number = $3,
			github_pull_request_url = $4,
			github_organization_id = $5
		WHERE github_pull_request_node_id = $6;`,
		pullRequest.RepositoryOwner,
		pullRequest.RepositoryName,
		pullRequest.Number,
		pullRequest.URL,
		pullRequest.OrganizationID,
		pullRequest.NodeID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (g GithubPullRequest) DeletePullRequestByInternalTaskID(
	ct context.Context,
	internalTaskID uint64,
) *errs.Error {
	_, err := g.db.Exec(`
		DELETE FROM apps_github_pull_request
		WHERE internal_task_id = $1;
		`,
		internalTaskID)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (g GithubPullRequest) DeletePullRequestByGithubNodeID(
	ct context.Context,
	githubNodeID string,
) *errs.Error {
	_, err := g.db.Exec(`
		DELETE FROM apps_github_pull_request
		WHERE github_pull_request_node_id = $1;
		`,
		githubNodeID)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (g GithubPullRequest) FindAllPullRequests(ct context.Context) ([]entity.GithubPullRequest, *errs.Error) {
	rows, err := g.db.Query(`
	SELECT
	    internal_task_id,
	    github_pull_request_node_id,
	 	github_repository_owner,
	    github_repository_name,
	    github_pull_request_number,
	    github_pull_request_url,
	 	github_organization_id
	FROM apps_github_pull_request;`,
	)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	var internalErr *errs.Error
	var pullRequests []entity.GithubPullRequest
	for rows.Next() {
		var pullRequest entity.GithubPullRequest
		err = rows.Scan(
			&pullRequest.InternalTaskID,
			&pullRequest.NodeID,
			&pullRequest.RepositoryOwner,
			&pullRequest.RepositoryName,
			&pullRequest.Number,
			&pullRequest.URL,
			&pullRequest.OrganizationID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		pullRequests = append(pullRequests, pullRequest)
	}

	return pullRequests, internalErr
}

func NewGithubPullRequest(dataCollector telemetry.DataCollector, sqlDB *sql.DB) GithubPullRequest {
	return GithubPullRequest{dataCollector: dataCollector, db: sqlDB}
}
