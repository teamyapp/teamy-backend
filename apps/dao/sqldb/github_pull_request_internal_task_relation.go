package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubPullRequestInternalTaskRelation struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.GithubPullRequestInternalTaskRelation = (*GithubPullRequestInternalTaskRelation)(nil)

func (g GithubPullRequestInternalTaskRelation) FindPullRequestInternalTaskRelationsByInternalTaskID(ct context.Context, internalTaskID uint64) ([]entity.GithubPullRequestInternalTaskRelation, *errs.Error) {
	rows, err := g.db.Query(`
	SELECT
		internal_task_id,
		pull_request_node_id,
		automatic_tracking,
		pull_request_task_link_id
	FROM apps_github_pull_request_internal_task_relation
	WHERE internal_task_id = $1;`, internalTaskID)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	defer rows.Close()
	var internalErr *errs.Error
	githubPullRequestInternalTaskRelations := make([]entity.GithubPullRequestInternalTaskRelation, 0)
	for rows.Next() {
		githubPullRequestInternalTaskRelation := entity.GithubPullRequestInternalTaskRelation{}
		err = rows.Scan(
			&githubPullRequestInternalTaskRelation.InternalTaskID,
			&githubPullRequestInternalTaskRelation.PullRequestNodeID,
			&githubPullRequestInternalTaskRelation.AutomaticTracking,
			&githubPullRequestInternalTaskRelation.InternalTaskLinkID,
		)
		if err != nil {
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			g.dataCollector.Logger.ErrorWithContext(ct, newInternalErr)
			continue
		}

		githubPullRequestInternalTaskRelations = append(githubPullRequestInternalTaskRelations, githubPullRequestInternalTaskRelation)
	}

	return githubPullRequestInternalTaskRelations, internalErr
}

func (g GithubPullRequestInternalTaskRelation) FindPullRequestInternalTaskRelationsByNodeID(ct context.Context, nodeID string) ([]entity.GithubPullRequestInternalTaskRelation, *errs.Error) {
	rows, err := g.db.Query(`
	SELECT
		internal_task_id,
		pull_request_node_id,
		automatic_tracking,
		pull_request_task_link_id
	FROM apps_github_pull_request_internal_task_relation
	WHERE pull_request_node_id = $1;`, nodeID)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	defer rows.Close()
	var internalErr *errs.Error
	githubPullRequestInternalTaskRelations := make([]entity.GithubPullRequestInternalTaskRelation, 0)
	for rows.Next() {
		githubPullRequestInternalTaskRelation := entity.GithubPullRequestInternalTaskRelation{}
		err = rows.Scan(
			&githubPullRequestInternalTaskRelation.InternalTaskID,
			&githubPullRequestInternalTaskRelation.PullRequestNodeID,
			&githubPullRequestInternalTaskRelation.AutomaticTracking,
			&githubPullRequestInternalTaskRelation.InternalTaskLinkID,
		)
		if err != nil {
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			g.dataCollector.Logger.ErrorWithContext(ct, newInternalErr)
			continue
		}

		githubPullRequestInternalTaskRelations = append(githubPullRequestInternalTaskRelations, githubPullRequestInternalTaskRelation)
	}

	return githubPullRequestInternalTaskRelations, internalErr
}

func (g GithubPullRequestInternalTaskRelation) CreatePullRequestInternalTaskRelation(ct context.Context, pullRequestInternalTaskRelation entity.GithubPullRequestInternalTaskRelation) *errs.Error {
	_, err := g.db.Exec(`
	INSERT INTO apps_github_pull_request_internal_task_relation
	(
	    internal_task_id,
	    pull_request_node_id,
		automatic_tracking,
	    pull_request_task_link_id
	)
	VALUES ($1, $2, $3, $4);
`,
		pullRequestInternalTaskRelation.InternalTaskID,
		pullRequestInternalTaskRelation.PullRequestNodeID,
		pullRequestInternalTaskRelation.AutomaticTracking,
		pullRequestInternalTaskRelation.InternalTaskLinkID,
	)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (g GithubPullRequestInternalTaskRelation) DeletePullRequestInternalTaskRelationByNodeIDAndTaskID(ct context.Context, nodeID string, internalTaskID uint64) *errs.Error {
	_, err := g.db.Exec(`
		DELETE FROM apps_github_pull_request_internal_task_relation
		WHERE internal_task_id = $1 AND pull_request_node_id = $2;
		`,
		internalTaskID, nodeID)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func NewGithubPullRequestInternalTaskRelation(dataCollector telemetry.DataCollector, sqlDB *sql.DB) GithubPullRequestInternalTaskRelation {
	return GithubPullRequestInternalTaskRelation{dataCollector: dataCollector, db: sqlDB}
}
