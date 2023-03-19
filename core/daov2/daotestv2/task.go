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

type Task struct {
	db *dbtest.InMemoryDB
}

var _ daov2.Task = (*Task)(nil)

func (t Task) FindTaskByID(ct context.Context, taskID uint64) (entity.Task, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t Task) FindTasksByTeamID(ct context.Context, teamID uint64) ([]entity.Task, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t Task) FindAllTasks(ct context.Context) ([]entity.Task, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t Task) FindTaskByIDWithTx(ct context.Context, tx *transaction.Transaction, taskID uint64) (entity.Task, *errs.Error) {
	var task entity.Task
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TaskTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currTask := rawRow.(entity.Task)
				if currTask.ID == taskID {
					task = currTask
					return nil
				}
			}

			return &errs.Error{
				Code:    errs.NotFound,
				Message: fmt.Sprintf("row not found: taskID=%v", taskID),
			}
		},
	})
	return task, err
}

func (t Task) FindTaskByCommentsThreadIDWithTx(ct context.Context, tx *transaction.Transaction, commentThreadID uint64) (entity.Task, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t Task) FindTasksByIDsWithTx(ct context.Context, tx *transaction.Transaction, taskIDs []uint64) ([]entity.Task, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t Task) FindTasksByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.Task, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t Task) FindAllTasksWithTx(ct context.Context, tx *transaction.Transaction) ([]entity.Task, *errs.Error) {
	//TODO implement me
	panic("implement me")
}

func (t Task) CreateTask(ct context.Context, tx *transaction.Transaction, task entity.Task) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TaskTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currTask := row.(entity.Task)
				if currTask.ID == task.ID {
					return &errs.Error{
						Code:    errs.AlreadyExists,
						Message: fmt.Sprintf("row already exist: taskID=%v", task.ID),
					}
				}
			}

			table.Rows = append(table.Rows, task)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TaskTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currTask := row.(entity.Task)
				if currTask.ID == task.ID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (t Task) UpdateTask(ct context.Context, tx *transaction.Transaction, task entity.Task) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func (t Task) DeleteTask(ct context.Context, tx *transaction.Transaction, taskID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewTask(db *dbtest.InMemoryDB) Task {
	return Task{db: db}
}
