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
	logger telemetry.Logger
	db     *sql.DB
}

var _ dao.GithubCodeReview = (*GithubCodeReview)(nil)

func (g GithubCodeReview) FindCodeReviewByGithubReviewerID(
	ct context.Context,
	githubPullRequestNodeID string,
	githubReviewerNodeID string,
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
	WHERE github_pull_request_node_id=$1 AND github_reviewer_node_id = $2;
`,
		githubPullRequestNodeID,
		githubReviewerNodeID).
		Scan(
			&codeReview.GithubPullRequestNodeID,
			&codeReview.GithubReviewerNodeID,
			&codeReview.InternalCodeReviewTaskID,
			&codeReview.InternalAddressFeedbackTaskID,
			&codeReview.Round,
		)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.GithubCodeReview{}, errs.NewError(
				errs.NotFound,
				fmt.Sprintf(
					"GithubCodeReview not found: githubReviewerID=%v",
					githubReviewerID))
		}

		return entity.GithubCodeReview{}, errs.NewError(errs.Unknown, err.Error())
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

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.GithubCodeReview{}, errs.NewError(
				errs.NotFound,
				fmt.Sprintf(
					"GithubCodeReview not found: internalTaskID=%v",
					internalTaskID))
		}

		return entity.GithubCodeReview{}, errs.NewError(errs.Unknown, err.Error())
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
		return errs.NewError(errs.Unknown, err.Error())
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
		return errs.NewError(errs.Unknown, err.Error())
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
		return errs.NewError(errs.Unknown, err.Error())
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
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewGithubCodeReview(logger telemetry.Logger, sqlDB *sql.DB) GithubCodeReview {
	return GithubCodeReview{logger: logger, db: sqlDB}
}
