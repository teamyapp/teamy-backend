package daotestv2

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
)

type Thread struct {
	db *dbtest.InMemoryDB
}

var _ daov2.Thread = (*Thread)(nil)

func (t Thread) CreateThread(ct context.Context, tx *transaction.Transaction, threadID uint64) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(ThreadTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currThreadID := row.(uint64)
				if currThreadID == threadID {
					return &errs.Error{
						Code:    errs.AlreadyExists,
						Message: fmt.Sprintf("row already exist: threadID=%v", threadID),
					}
				}
			}

			table.Rows = append(table.Rows, threadID)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(ThreadTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currThreadID := row.(uint64)
				if currThreadID == threadID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (t Thread) DeleteThread(ct context.Context, tx *transaction.Transaction, threadID uint64) *errs.Error {
	//TODO implement me
	panic("implement me")
}

func NewThread(db *dbtest.InMemoryDB) Thread {
	return Thread{db: db}
}
