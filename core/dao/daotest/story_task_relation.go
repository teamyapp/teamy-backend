package daotest

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type StoryTaskRelation struct {
	db *dbtest.InMemoryDB
}

func (s StoryTaskRelation) FindTaskIDsByStoryIDWithTx(ct context.Context, tx *transaction.Transaction, storyID uint64) ([]uint64, *errs.Error) {
	var taskIDs []uint64
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(StoryTaskRelationTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currStoryTaskRelation := rawRow.(entity.StoryTaskRelation)
				if currStoryTaskRelation.StoryID == storyID {
					taskIDs = append(taskIDs, currStoryTaskRelation.TaskID)
				}
			}

			return nil
		},
	})
	return taskIDs, err
}

func (s StoryTaskRelation) FindStoryIDsByTaskIDWithTx(ct context.Context, tx *transaction.Transaction, taskID uint64) ([]uint64, *errs.Error) {
	var storyIDs []uint64
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(StoryTaskRelationTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currStoryTaskRelation := rawRow.(entity.StoryTaskRelation)
				if currStoryTaskRelation.TaskID == taskID {
					storyIDs = append(storyIDs, currStoryTaskRelation.StoryID)
				}
			}

			return nil
		},
	})
	return storyIDs, err
}

func (s StoryTaskRelation) CreateStoryTaskRelation(ct context.Context, tx *transaction.Transaction, storyTaskRelation entity.StoryTaskRelation) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(StoryTaskRelationTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currStoryTaskRelation := rawRow.(entity.StoryTaskRelation)
				if currStoryTaskRelation.StoryID == storyTaskRelation.StoryID &&
					currStoryTaskRelation.TaskID == storyTaskRelation.TaskID {
					return &errs.Error{
						Code:    errs.AlreadyExists,
						Message: fmt.Sprintf("row already exist: storyId=%v taskId=%v", storyTaskRelation.StoryID, storyTaskRelation.TaskID),
					}
				}
			}

			table.Rows = append(table.Rows, storyTaskRelation)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := s.db.GetTable(StoryTaskRelationTableName)
			if err != nil {
				return err
			}

			for index, rawRow := range table.Rows {
				currStoryTaskRelation := rawRow.(entity.StoryTaskRelation)
				if currStoryTaskRelation.StoryID == storyTaskRelation.StoryID &&
					currStoryTaskRelation.TaskID == storyTaskRelation.TaskID {
					table.Rows = append(table.Rows[:index], table.Rows[index+1:]...)
					return nil
				}
			}

			return &errs.Error{
				Code:    errs.NotFound,
				Message: fmt.Sprintf("row not found: storyId=%v taskId=%v", storyTaskRelation.StoryID, storyTaskRelation.TaskID),
			}
		},
	})
}

func (s StoryTaskRelation) DeleteStoryTaskRelation(ct context.Context, tx *transaction.Transaction, storyID uint64, taskID uint64) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(StoryTaskRelationTableName)
			if err != nil {
				return err
			}

			for index, rawRow := range table.Rows {
				currStoryTaskRelation := rawRow.(entity.StoryTaskRelation)
				if currStoryTaskRelation.StoryID == storyID &&
					currStoryTaskRelation.TaskID == taskID {
					table.Rows = append(table.Rows[:index], table.Rows[index+1:]...)
					return nil
				}
			}

			return &errs.Error{
				Code:    errs.NotFound,
				Message: fmt.Sprintf("row not found: storyId=%v taskId=%v", storyID, taskID),
			}
		},
		Undo: func() *errs.Error {
			table, err := s.db.GetTable(StoryTaskRelationTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currStoryTaskRelation := rawRow.(entity.StoryTaskRelation)
				if currStoryTaskRelation.StoryID == storyID &&
					currStoryTaskRelation.TaskID == taskID {
					return &errs.Error{
						Code:    errs.AlreadyExists,
						Message: fmt.Sprintf("row already exist: storyId=%v taskId=%v", storyID, taskID),
					}
				}
			}

			table.Rows = append(table.Rows, entity.StoryTaskRelation{
				StoryID: storyID,
				TaskID:  taskID,
			})
			return nil
		},
	})
}

func (s StoryTaskRelation) DeleteStoryTaskRelationsByStoryID(ct context.Context, tx *transaction.Transaction, storyID uint64) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(StoryTaskRelationTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, rawRow := range table.Rows {
				currStoryTaskRelation := rawRow.(entity.StoryTaskRelation)
				if currStoryTaskRelation.StoryID == storyID {
					continue
				}

				rows = append(rows, currStoryTaskRelation)
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := s.db.GetTable(StoryTaskRelationTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, rawRow := range table.Rows {
				currStoryTaskRelation := rawRow.(entity.StoryTaskRelation)
				if currStoryTaskRelation.StoryID == storyID {
					rows = append(rows, currStoryTaskRelation)
				}
			}

			table.Rows = rows
			return nil
		},
	})
}

func (s StoryTaskRelation) DeleteStoryTaskRelationsByTaskID(ct context.Context, tx *transaction.Transaction, taskID uint64) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(StoryTaskRelationTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, rawRow := range table.Rows {
				currStoryTaskRelation := rawRow.(entity.StoryTaskRelation)
				if currStoryTaskRelation.TaskID == taskID {
					continue
				}

				rows = append(rows, currStoryTaskRelation)
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := s.db.GetTable(StoryTaskRelationTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, rawRow := range table.Rows {
				currStoryTaskRelation := rawRow.(entity.StoryTaskRelation)
				if currStoryTaskRelation.TaskID == taskID {
					rows = append(rows, currStoryTaskRelation)
				}
			}

			table.Rows = rows
			return nil
		},
	})
}

func NewStoryTaskRelation(db *dbtest.InMemoryDB) StoryTaskRelation {
	return StoryTaskRelation{db: db}
}
