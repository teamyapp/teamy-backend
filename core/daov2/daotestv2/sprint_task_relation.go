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

type SprintTaskRelation struct {
	db *dbtest.InMemoryDB
}

var _ daov2.SprintTaskRelation = (*SprintTaskRelation)(nil)

func (s SprintTaskRelation) FindTaskIDsBySprintIDWithTx(ct context.Context, tx *transaction.Transaction, sprintID uint64) ([]uint64, *errs.Error) {
	var taskIDs []uint64
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(SprintTaskRelationTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currSprintTaskRelation := rawRow.(entity.SprintTaskRelation)
				if currSprintTaskRelation.SprintID == sprintID {
					taskIDs = append(taskIDs, currSprintTaskRelation.TaskID)
				}
			}

			return nil
		},
	})
	return taskIDs, err
}

func (s SprintTaskRelation) FindSprintIDsByTaskIDWithTx(ct context.Context, tx *transaction.Transaction, taskID uint64) ([]uint64, *errs.Error) {
	var sprintIDs []uint64
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(SprintTaskRelationTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currSprintTaskRelation := rawRow.(entity.SprintTaskRelation)
				if currSprintTaskRelation.TaskID == taskID {
					sprintIDs = append(sprintIDs, currSprintTaskRelation.SprintID)
				}
			}

			return nil
		},
	})
	return sprintIDs, err
}

func (s SprintTaskRelation) CreateSprintTaskRelation(ct context.Context, tx *transaction.Transaction, relation entity.SprintTaskRelation) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(SprintTaskRelationTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currSprintTaskRelation := row.(entity.SprintTaskRelation)
				if currSprintTaskRelation.SprintID == relation.SprintID && currSprintTaskRelation.TaskID == relation.TaskID {
					return &errs.Error{
						Code:    errs.AlreadyExists,
						Message: fmt.Sprintf("row already exist: sprintId=%v taskId=%v", relation.SprintID, relation.TaskID),
					}
				}
			}

			table.Rows = append(table.Rows, relation)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := s.db.GetTable(SprintTaskRelationTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currSprintTaskRelation := row.(entity.SprintTaskRelation)
				if currSprintTaskRelation.TaskID == relation.TaskID && currSprintTaskRelation.SprintID == relation.SprintID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (s SprintTaskRelation) DeleteSprintTaskRelation(ct context.Context, tx *transaction.Transaction, sprintID uint64, taskID uint64) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(SprintTaskRelationTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currSprintTaskRelation := row.(entity.SprintTaskRelation)
				if currSprintTaskRelation.TaskID == taskID && currSprintTaskRelation.SprintID == sprintID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := s.db.GetTable(SprintTaskRelationTableName)
			if err != nil {
				return err
			}

			table.Rows = append(table.Rows, entity.SprintTaskRelation{
				SprintID: sprintID,
				TaskID:   taskID,
			})
			return nil
		},
	})
}

func NewSprintTaskRelation(db *dbtest.InMemoryDB) SprintTaskRelation {
	return SprintTaskRelation{db: db}
}
