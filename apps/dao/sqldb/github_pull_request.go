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

func (g GithubPullRequest) FindPullRequestByInternalTaskID(
	ct context.Context,
	internalTaskID uint64,
) (entity.GithubPullRequest, *errs.Error) {
	pullRequest := entity.GithubPullRequest{}
	err := g.db.QueryRow(`
	SELECT
	    internal_task_id,
	    github_pull_request_node_id,
	    github_repository_owner,
	    github_repository_name,
	    github_pull_request_number,
	    github_pull_request_url
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
		internalErr := &errs.Error{
			Code: errs.NotFound,
			Message: fmt.Sprintf(
				"GithubPullRequest not found: internalTaskID=%v", internalTaskID),
		}
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.GithubPullRequest{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.GithubPullRequest{}, internalErr
	}

	return pullRequest, nil
}

func (g GithubPullRequest) FindPullRequestByGithubNodeID(
	ct context.Context,
	githubNodeID string,
) (entity.GithubPullRequest, *errs.Error) {
	pullRequest := entity.GithubPullRequest{}
	err := g.db.QueryRow(`
	SELECT
	    internal_task_id,
	    github_pull_request_node_id,
	    github_repository_owner,
	    github_repository_name,
	    github_pull_request_number,
	    github_pull_request_url
	FROM apps_github_pull_request
	WHERE github_pull_request_node_id = $1;
`,
		githubNodeID).
		Scan(
			&pullRequest.InternalTaskID,
			&pullRequest.NodeID,
			&pullRequest.RepositoryOwner,
			&pullRequest.RepositoryName,
			&pullRequest.Number,
			&pullRequest.URL,
		)

	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code: errs.NotFound,
			Message: fmt.Sprintf(fmt.Sprintf(
				"GithubPullRequest not found: githubNodeID=%v",
				githubNodeID)),
		}
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.GithubPullRequest{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.GithubPullRequest{}, internalErr
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
	    internal_task_id,
	    github_pull_request_node_id,
	 	github_repository_owner,
	    github_repository_name,
	    github_pull_request_number,
	    github_pull_request_url
	)
	VALUES ($1, $2, $3, $4, $5, $6);
`,
		pullRequest.InternalTaskID,
		pullRequest.NodeID,
		pullRequest.RepositoryOwner,
		pullRequest.RepositoryName,
		pullRequest.Number,
		pullRequest.URL,
	)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func NewGithubPullRequest(dataCollector telemetry.DataCollector, sqlDB *sql.DB) GithubPullRequest {
	return GithubPullRequest{dataCollector: dataCollector, db: sqlDB}
}
