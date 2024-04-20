package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const storyTaskRelationDaoName = "StoryTaskRelation"

type StoryTaskRelation struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.StoryTaskRelation = (*StoryTaskRelation)(nil)

func (s *StoryTaskRelation) FindTaskIDsByStoryIDWithTx(ct context.Context, tx *transaction.Transaction, storyID uint64) ([]uint64, *errs.Error) {
	s.metrics.ReportDaoOperation(storyTaskRelationDaoName, "FindTaskIDsByStoryIDWithTx")
	var taskIDs []uint64
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

func (s *StoryTaskRelation) FindStoryIDsByTaskIDWithTx(ct context.Context, tx *transaction.Transaction, taskID uint64) ([]uint64, *errs.Error) {
	s.metrics.ReportDaoOperation(storyTaskRelationDaoName, "FindStoryIDsByTaskIDWithTx")
	var storyIDs []uint64
	rows, err := tx.SQLTx().QueryContext(ct, `
		SELECT
			story_id
		FROM story_task_relation
		WHERE task_id = $1
	`, taskID)
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

func (s *StoryTaskRelation) CreateStoryTaskRelation(ct context.Context, tx *transaction.Transaction, storyTaskRelation entity.StoryTaskRelation) *errs.Error {
	s.metrics.ReportDaoOperation(storyTaskRelationDaoName, "CreateStoryTaskRelation")
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
	s.metrics.ReportDaoOperation(storyTaskRelationDaoName, "DeleteStoryTaskRelation")
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM story_task_relation
		WHERE story_id = $1 AND task_id = $2
	`, storyID, taskID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (s *StoryTaskRelation) DeleteStoryTaskRelationsByStoryID(ct context.Context, tx *transaction.Transaction, storyID uint64) *errs.Error {
	s.metrics.ReportDaoOperation(storyTaskRelationDaoName, "DeleteStoryTaskRelationsByStoryID")
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM story_task_relation
		WHERE story_id = $1
	`, storyID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (s *StoryTaskRelation) DeleteStoryTaskRelationsByTaskID(ct context.Context, tx *transaction.Transaction, taskID uint64) *errs.Error {
	s.metrics.ReportDaoOperation(storyTaskRelationDaoName, "DeleteStoryTaskRelationsByTaskID")
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM story_task_relation
		WHERE task_id = $1
	`, taskID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewStoryTaskRelation(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *StoryTaskRelation {
	return &StoryTaskRelation{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
