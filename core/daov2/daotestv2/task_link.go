package daotestv2

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskLink struct {
	db *dbtest.InMemoryDB
}

var _ daov2.TaskLink = (*TaskLink)(nil)

func (t TaskLink) FindTaskLinkByID(ct context.Context, tx *transaction.Transaction, taskLinkID uint64) (entity.TaskLink, *errs.Error) {
	var taskLink entity.TaskLink
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TaskLinkTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currTaskLink := rawRow.(entity.TaskLink)
				if currTaskLink.ID == taskLinkID {
					taskLink = currTaskLink
					return nil
				}
			}

			return &errs.Error{
				Code:    errs.NotFound,
				Message: fmt.Sprintf("row not found: taskLinkID=%v", taskLinkID),
			}
		},
	})
	return taskLink, err
}

func (t TaskLink) FindLinksByTaskID(ct context.Context, tx *transaction.Transaction, taskID uint64) ([]entity.TaskLink, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t TaskLink) CreateTaskLink(ct context.Context, tx *transaction.Transaction, taskLinkEntity entity.TaskLink) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TaskLinkTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currTaskLink := row.(entity.TaskLink)
				if currTaskLink.ID == taskLinkEntity.ID {
					return &errs.Error{
						Code:    errs.AlreadyExists,
						Message: fmt.Sprintf("row already exist: taskLinkID=%v", taskLinkEntity.ID),
					}
				}
			}

			table.Rows = append(table.Rows, taskLinkEntity)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TaskLinkTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currTaskLink := row.(entity.TaskLink)
				if currTaskLink.ID == taskLinkEntity.ID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (t TaskLink) DeleteTaskLink(ct context.Context, tx *transaction.Transaction, taskLinkID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewTaskLink(db *dbtest.InMemoryDB) TaskLink {
	return TaskLink{db: db}
}
