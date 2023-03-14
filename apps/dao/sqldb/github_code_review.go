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

type GithubCodeReview struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.GithubCodeReview = (*GithubCodeReview)(nil)

func (g GithubCodeReview) FindCodeReviewByGithubReviewerID(
	ct context.Context,
	githubPullRequestNodeID string,
	githubReviewerID string,
) (entity.GithubCodeReview, *errs.Error) {
	codeReview := entity.GithubCodeReview{}
	err := g.db.QueryRow(`
	SELECT
	    github_pull_request_node_id,
    	github_reviewer_node_id,
    	internal_code_review_task_id,
    	internal_address_feedback_task_id,
    	round
	FROM apps_github_code_review
	WHERE github_pull_request_node_id=$1 AND github_reviewer_id = $2;
`,
		githubPullRequestNodeID,
		githubReviewerID).
		Scan(
			&codeReview.GithubPullRequestNodeID,
			&codeReview.GithubReviewerNodeID,
			&codeReview.InternalCodeReviewTaskID,
			&codeReview.InternalAddressFeedbackTaskID,
			&codeReview.Round,
		)

	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code: errs.NotFound,
			Message: fmt.Sprintf(
				"GithubCodeReview not found: githubReviewerID=%v",
				githubReviewerID),
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.GithubCodeReview{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.GithubCodeReview{}, internalErr
	}

	return codeReview, nil
}

func (g GithubCodeReview) FindCodeReviewByInternalTaskID(
	ct context.Context,
	internalTaskID uint64,
) (entity.GithubCodeReview, *errs.Error) {
	codeReview := entity.GithubCodeReview{}
	err := g.db.QueryRow(`
	SELECT
	    github_pull_request_node_id,
    	github_reviewer_node_id,
    	internal_code_review_task_id,
    	internal_address_feedback_task_id,
    	round
	FROM apps_github_code_review
	WHERE internal_task_id = $1;
`,
		internalTaskID).
		Scan(
			&codeReview.GithubPullRequestNodeID,
			&codeReview.GithubReviewerNodeID,
			&codeReview.InternalCodeReviewTaskID,
			&codeReview.InternalAddressFeedbackTaskID,
			&codeReview.Round,
		)
	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code: errs.NotFound,
			Message: fmt.Sprintf(
				"GithubCodeReview not found: internalTaskID=%v",
				internalTaskID),
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.GithubCodeReview{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.GithubCodeReview{}, internalErr
	}

	return codeReview, nil
}

func (g GithubCodeReview) CreateCodeReview(
	ct context.Context,
	codeReview entity.GithubCodeReview,
) *errs.Error {
	_, err := g.db.Exec(`
	INSERT INTO apps_github_code_review
	(
	    github_pull_request_node_id,
    	github_reviewer_node_id,
    	internal_code_review_task_id,
    	internal_address_feedback_task_id,
    	round
	)
	VALUES ($1, $2, $3, $4, $5);
`,
		codeReview.GithubPullRequestNodeID,
		codeReview.GithubReviewerNodeID,
		codeReview.InternalCodeReviewTaskID,
		codeReview.InternalAddressFeedbackTaskID,
		codeReview.Round,
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

func (g GithubCodeReview) UpdateCodeReview(
	ct context.Context,
	codeReview entity.GithubCodeReview,
) *errs.Error {
	_, err := g.db.Exec(`
		UPDATE apps_github_code_review
		SET
			github_pull_request_node_id = $1,
    		github_reviewer_node_id = $2,
			internal_code_review_task_id = $3,
			internal_address_feedback_task_id = $4,
			round = $5
		WHERE github_pull_request_node_id=$6 AND github_reviewer_node_id = $7;`,
		codeReview.GithubPullRequestNodeID,
		codeReview.GithubReviewerNodeID,
		codeReview.InternalCodeReviewTaskID,
		codeReview.InternalAddressFeedbackTaskID,
		codeReview.Round,
		codeReview.GithubPullRequestNodeID,
		codeReview.GithubReviewerNodeID,
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

func (g GithubCodeReview) DeleteCodeReviewByInternalTaskID(
	ct context.Context,
	internalTaskID uint64,
) *errs.Error {
	_, err := g.db.Exec(`
		DELETE FROM apps_github_code_review
		WHERE internal_task_id = $1;
		`,
		internalTaskID)
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

func (g GithubCodeReview) DeleteCodeReviewByGithubReviewerID(
	ct context.Context,
	githubReviewerID uint64,
) *errs.Error {
	_, err := g.db.Exec(`
		DELETE FROM apps_github_code_review
		WHERE github_reviewer_id = $1;
		`,
		githubReviewerID)
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

func NewGithubCodeReview(dataCollector telemetry.DataCollector, sqlDB *sql.DB) GithubCodeReview {
	return GithubCodeReview{dataCollector: dataCollector, db: sqlDB}
}
