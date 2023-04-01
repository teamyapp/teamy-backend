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

type Team struct {
	db                 *dbtest.InMemoryDB
	transactionFactory transaction.Factory
}

var _ daov2.Team = (*Team)(nil)

func (t Team) FindTeamByID(ct context.Context, teamID uint64) (entity.Team, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := t.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.Team{}, err
	}

	defer tx.Rollback()
	return t.FindTeamByIDWithTx(ct, tx, teamID)
}

func (t Team) FindAllTeams(ct context.Context) ([]entity.Team, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := t.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return t.FindAllTeamsWithTx(ct, tx)
}

func (t Team) FindAllTeamsWithTx(ct context.Context, tx *transaction.Transaction) ([]entity.Team, *errs.Error) {
	var teams []entity.Team
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currTeam := rawRow.(entity.Team)
				teams = append(teams, currTeam)
			}

			return nil
		},
	})
	return teams, err
}

func (t Team) FindTeamByIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) (entity.Team, *errs.Error) {
	var team entity.Team
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currTeam := rawRow.(entity.Team)
				if currTeam.ID == teamID {
					team = currTeam
					return nil
				}
			}

			return &errs.Error{
				Code:    errs.NotFound,
				Message: fmt.Sprintf("row not found: teamID=%v", teamID),
			}
		},
	})
	return team, err
}

func (t Team) FindTeamsByIDsWithTx(ct context.Context, tx *transaction.Transaction, teamIDs []uint64) ([]entity.Team, *errs.Error) {
	var teams []entity.Team
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamTableName)
			if err != nil {
				return err
			}

			teamMap := make(map[uint64]bool)
			for _, teamID := range teamIDs {
				teamMap[teamID] = true
			}

			for _, rawRow := range table.Rows {
				currTeam := rawRow.(entity.Team)
				if _, ok := teamMap[currTeam.ID]; ok {
					teams = append(teams, currTeam)
				}
			}

			return nil
		},
	})
	return teams, err
}

func (t Team) CreateTeam(ct context.Context, tx *transaction.Transaction, team entity.Team) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				currTeam := row.(entity.Team)
				if currTeam.ID == team.ID {
					return errs.NewError(errs.Unknown, fmt.Sprintf("row already exist: teamID=%v", team.ID))
				}
			}

			table.Rows = append(table.Rows, team)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TeamTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currTeam := row.(entity.Team)
				if currTeam.ID == team.ID {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (t Team) UpdateTeam(ct context.Context, tx *transaction.Transaction, team entity.Team) *errs.Error {
	oldTeam, internalErr := t.FindTeamByIDWithTx(ct, tx, team.ID)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamTableName)
			if err != nil {
				return err
			}

			for i, row := range table.Rows {
				currTeam := row.(entity.Team)
				if currTeam.ID == team.ID {
					table.Rows[i] = team
					return nil
				}
			}

			return errs.NewError(errs.Unknown, fmt.Sprintf("row not exist: teamID=%v", team.ID))
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TeamTableName)
			if err != nil {
				return err
			}

			for index, row := range table.Rows {
				currTeam := row.(entity.Team)
				if currTeam.ID == team.ID {
					table.Rows[index] = oldTeam
				}
			}

			return nil
		},
	})
}

func (t Team) DeleteTeam(ct context.Context, tx *transaction.Transaction, teamID uint64) *errs.Error {
	oldTeam, internalErr := t.FindTeamByIDWithTx(ct, tx, teamID)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currTeam := row.(entity.Team)
				if currTeam.ID != teamID {
					rows = append(rows, currTeam)
				}
			}


			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TeamTableName)
			if err != nil {
				return err
			}

			table.Rows = append(table.Rows, oldTeam)
			return nil
		},
	})
}

func NewTeam(db *dbtest.InMemoryDB, transactionFactory transaction.Factory) Team {
	return Team{db: db, transactionFactory: transactionFactory}
}
