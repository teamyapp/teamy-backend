package daotest

import (
	"context"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TaskFileUploadSession struct {
	db *dbtest.InMemoryDB
}

var _ dao.TaskFileUploadSession = (*TaskFileUploadSession)(nil)

func (t *TaskFileUploadSession) FindTaskFileUploadSessionByTaskIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	taskID uint64,
	taskFileUploadSessionType entity.TaskFileUploadSessionType,
	fileUploadSessionID uint64,
) (entity.TaskFileUploadSession, *errs.Error) {
	var taskFileUploadSession entity.TaskFileUploadSession

	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TaskFileUploadSessionTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currTaskFileUploadSession := rawRow.(entity.TaskFileUploadSession)
				if currTaskFileUploadSession.FileUploadSessionID == fileUploadSessionID && currTaskFileUploadSession.TaskID == taskID && currTaskFileUploadSession.Type == taskFileUploadSessionType {
					taskFileUploadSession = currTaskFileUploadSession
					break
				}
			}

			return nil
		},
	})

	return taskFileUploadSession, err
}

func (t *TaskFileUploadSession) CreateTaskFileUploadSession(ct context.Context, tx *transaction.Transaction, taskFileUploadSession entity.TaskFileUploadSession) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TaskFileUploadSessionTableName)
			if err != nil {
				return err
			}

			table.Rows = append(table.Rows, taskFileUploadSession)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TaskFileUploadSessionTableName)
			if err != nil {
				return err
			}

			for i, rawRow := range table.Rows {
				currTaskFileUploadSession := rawRow.(entity.TaskFileUploadSession)
				if currTaskFileUploadSession.FileUploadSessionID == taskFileUploadSession.FileUploadSessionID && currTaskFileUploadSession.TaskID == taskFileUploadSession.TaskID && currTaskFileUploadSession.Type == taskFileUploadSession.Type {
					table.Rows = append(table.Rows[:i], table.Rows[i+1:]...)
					break
				}
			}

			return nil
		},
	})
}

func (t *TaskFileUploadSession) UpdateTaskFileUploadSession(ct context.Context, tx *transaction.Transaction, taskFileUploadSession entity.TaskFileUploadSession) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TaskFileUploadSessionTableName)
			if err != nil {
				return err
			}

			for i, rawRow := range table.Rows {
				currTaskFileUploadSession := rawRow.(entity.TaskFileUploadSession)
				if currTaskFileUploadSession.FileUploadSessionID == taskFileUploadSession.FileUploadSessionID && currTaskFileUploadSession.TaskID == taskFileUploadSession.TaskID && currTaskFileUploadSession.Type == taskFileUploadSession.Type {
					table.Rows[i] = taskFileUploadSession
					break
				}
			}

			return nil
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TaskFileUploadSessionTableName)
			if err != nil {
				return err
			}

			for i, rawRow := range table.Rows {
				currTaskFileUploadSession := rawRow.(entity.TaskFileUploadSession)
				if currTaskFileUploadSession.FileUploadSessionID == taskFileUploadSession.FileUploadSessionID && currTaskFileUploadSession.TaskID == taskFileUploadSession.TaskID && currTaskFileUploadSession.Type == taskFileUploadSession.Type {
					table.Rows[i] = currTaskFileUploadSession
					break
				}
			}

			return nil
		},
	})

}

func NewTaskFileUploadSession(db *dbtest.InMemoryDB) *TaskFileUploadSession {
	return &TaskFileUploadSession{
		db: db,
	}
}
