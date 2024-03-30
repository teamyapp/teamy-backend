package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type ProjectStoryRelation struct {
	transactionFactory transaction.Factory
}

var _ dao.ProjectStoryRelation = (*ProjectStoryRelation)(nil)

func (p *ProjectStoryRelation) FindStoryIDsByProjectIDWithTx(ct context.Context, tx *transaction.Transaction, projectID uint64) ([]uint64, *errs.Error) {
	storyIDs := []uint64{}
	rows, err := tx.SQLTx().QueryContext(ct, `
		SELECT
			story_id
		FROM project_story_relation
		WHERE project_id = $1
	`, projectID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		var storyID uint64
		err := rows.Scan(&storyID)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		storyIDs = append(storyIDs, storyID)
	}

	return storyIDs, nil
}

func (p *ProjectStoryRelation) CreateProjectStoryRelation(ct context.Context, tx *transaction.Transaction, projectStoryRelation entity.ProjectStoryRelation) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		INSERT INTO project_story_relation (
			project_id,
			story_id
		) VALUES (
			$1,
			$2
		)
	`,
		projectStoryRelation.ProjectID,
		projectStoryRelation.StoryID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (p *ProjectStoryRelation) DeleteProjectStoryRelation(ct context.Context, tx *transaction.Transaction, projectID uint64, storyID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM project_story_relation
		WHERE project_id = $1
		AND story_id = $2
	`, projectID, storyID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (p *ProjectStoryRelation) DeleteProjectStoryRelationsByProjectID(ct context.Context, tx *transaction.Transaction, projectID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM project_story_relation
		WHERE project_id = $1
	`, projectID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (p *ProjectStoryRelation) DeleteProjectStoryRelationsByStoryID(ct context.Context, tx *transaction.Transaction, storyID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM project_story_relation
		WHERE story_id = $1
	`, storyID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewProjectStoryRelation(transactionFactory transaction.Factory) *ProjectStoryRelation {
	return &ProjectStoryRelation{
		transactionFactory: transactionFactory,
	}
}
