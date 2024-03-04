package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type StoryTaskRelation struct {
	transactionFactory transaction.Factory
}

var _ dao.StoryTaskRelation = (*StoryTaskRelation)(nil)

func (s *StoryTaskRelation) FindTaskIDsByStoryIDWithTx(ct context.Context, tx *transaction.Transaction, storyID uint64) ([]uint64, *errs.Error) {
	taskIDs := []uint64{}
	rows, err := tx.SQLTx().QueryContext(ct, `
		SELECT
			task_id
		FROM story_task_relation
		WHERE story_id = $1
	`, storyID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		var taskID uint64
		err := rows.Scan(&taskID)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}
		
		taskIDs = append(taskIDs, taskID)
	}

	return taskIDs, nil
}

func (s *StoryTaskRelation) CreateStoryTaskRelation(ct context.Context, tx *transaction.Transaction, storyTaskRelation entity.StoryTaskRelation) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		INSERT INTO story_task_relation (
			story_id,
			task_id
		) VALUES (
			$1,
			$2
		)
	`,
		storyTaskRelation.StoryID,
		storyTaskRelation.TaskID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (s *StoryTaskRelation) DeleteStoryTaskRelation(ct context.Context, tx *transaction.Transaction, storyID uint64, taskID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM story_task_relation
		WHERE story_id = $1 AND task_id = $2
	`, storyID, taskID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewStoryTaskRelation(transactionFactory transaction.Factory) *StoryTaskRelation {
	return &StoryTaskRelation{
		transactionFactory: transactionFactory,
	}
}
