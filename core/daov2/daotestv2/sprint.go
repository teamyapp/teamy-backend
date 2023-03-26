package daotestv2

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/teamyapp/cloud/libs/dbtest"
	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Sprint struct {
	db                 *dbtest.InMemoryDB
	transactionFactory transaction.Factory
}

var _ daov2.Sprint = (*Sprint)(nil)

func (s Sprint) FindSprintByID(ct context.Context, sprintID uint64) (entity.Sprint, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := s.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.Sprint{}, err
	}

	defer tx.Rollback()
	return s.FindSprintByIDWithTx(ct, tx, sprintID)
}

func (s Sprint) FindSprintsByTeamID(ct context.Context, teamID uint64) ([]entity.Sprint, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := s.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return s.FindSprintsByTeamIDWithTx(ct, tx, teamID)
}

func (s Sprint) FindAllSprints(ct context.Context) ([]entity.Sprint, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := s.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return s.FindAllSprintsWithTx(ct, tx)
}

func (s Sprint) FindSprintByIDWithTx(ct context.Context, tx *transaction.Transaction, sprintID uint64) (entity.Sprint, *errs.Error) {
	var sprint entity.Sprint
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(SprintTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currSprint := rawRow.(entity.Sprint)
				if currSprint.ID == sprintID {
					sprint = currSprint
					return nil
				}
			}

			return &errs.Error{
				Code:    errs.NotFound,
				Message: fmt.Sprintf("row not found: sprintID=%v", sprintID),
			}
		},
	})
	return sprint, err
}

func (s Sprint) FindSprintsByIDsWithTx(ct context.Context, tx *transaction.Transaction, sprintIDs []uint64) ([]entity.Sprint, *errs.Error) {
	var sprints []entity.Sprint
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(SprintTableName)
			if err != nil {
				return err
			}

			sprintMap := make(map[uint64]int)
			for _, sprintID := range sprintIDs {
				sprintMap[sprintID]++
			}

			for _, rawRow := range table.Rows {
				currSprint := rawRow.(entity.Sprint)
				if _, ok := sprintMap[currSprint.ID]; ok {
					sprints = append(sprints, currSprint)
				}
			}

			return nil
		},
	})
	return sprints, err
}

func (s Sprint) FindSprintsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.Sprint, *errs.Error) {
	var sprints []entity.Sprint
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(SprintTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currSprint := rawRow.(entity.Sprint)
				if currSprint.OwningTeamID == teamID {
					sprints = append(sprints, currSprint)
				}
			}

			return nil
		},
	})
	return sprints, err
}

func (s Sprint) FindAllSprintsWithTx(ct context.Context, tx *transaction.Transaction) ([]entity.Sprint, *errs.Error) {
	var sprints []entity.Sprint
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(SprintTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currSprint := rawRow.(entity.Sprint)
				sprints = append(sprints, currSprint)
			}

			return nil
		},
	})
	return sprints, err
}

func (s Sprint) CreateSprint(ct context.Context, tx *transaction.Transaction, sprint entity.Sprint) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(SprintTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currSprint := row.(entity.Sprint)
				if currSprint.ID == sprint.ID {
					return errs.NewError(errs.Unknown, fmt.Sprintf("row already exist: sprintID=%v", sprint.ID))
				}
			}

			table.Rows = append(table.Rows, sprint)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := s.db.GetTable(SprintTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currSprint := row.(entity.Sprint)
				if currSprint.ID == sprint.ID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (s Sprint) DeleteSprint(ct context.Context, tx *transaction.Transaction, sprintID uint64) *errs.Error {
	oldSprint, internalErr := s.FindSprintByIDWithTx(ct, tx, sprintID)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := s.db.GetTable(SprintTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currSprint := row.(entity.Sprint)
				if currSprint.ID != sprintID {
					rows = append(rows, currSprint)
				}
			}

			if len(rows) == len(table.Rows) {
				return errs.NewError(errs.Unknown, fmt.Sprintf("row not exist: sprintID=%v", sprintID))
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := s.db.GetTable(SprintTableName)
			if err != nil {
				return err
			}

			table.Rows = append(table.Rows, oldSprint)
			return nil
		},
	})
}

func NewSprint(db *dbtest.InMemoryDB, transactionFactory transaction.Factory) Sprint {
	return Sprint{db: db, transactionFactory: transactionFactory}
}
