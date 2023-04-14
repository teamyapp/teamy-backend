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

type TeamGroup struct {
	db                 *dbtest.InMemoryDB
	transactionFactory transaction.Factory
}

var _ daov2.TeamGroup = (*TeamGroup)(nil)

func (t TeamGroup) FindGroupByTeamIDAndLabel(ct context.Context, teamID uint64, groupLabel string) (entity.TeamGroup, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := t.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return entity.TeamGroup{}, err
	}

	defer tx.Rollback()
	return t.FindGroupByTeamIDAndLabelWithTx(ct, tx, teamID, groupLabel)
}

func (t TeamGroup) FindGroupByTeamIDAndLabelWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	teamID uint64,
	groupLabel string,
) (entity.TeamGroup, *errs.Error) {
	var teamGroup entity.TeamGroup
	err := tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamGroupTableName)
			if err != nil {
				return err
			}

			for _, rawRow := range table.Rows {
				currGroup := rawRow.(entity.TeamGroup)
				if currGroup.TeamID == teamID &&
					currGroup.Label == groupLabel {
					teamGroup = currGroup
					return nil
				}
			}

			return &errs.Error{
				Code:    errs.NotFound,
				Message: fmt.Sprintf("row not found: teamID=%v, groupLabel=%v", teamID, groupLabel),
			}
		},
	})
	return teamGroup, err
}

func (t TeamGroup) CreateGroup(
	ct context.Context,
	tx *transaction.Transaction,
	group entity.TeamGroup,
) *errs.Error {
	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamGroupTableName)
			if err != nil {
				return err
			}

			for _, row := range table.Rows {
				teamGroup := row.(entity.TeamGroup)
				if teamGroup.TeamID == group.TeamID &&
					teamGroup.Label == group.Label {
					return errs.NewError(
						errs.Unknown,
						fmt.Sprintf("row already exist: teamID=%v, label=%v",
							teamGroup.TeamID,
							teamGroup.Label,
						))
				}
			}

			table.Rows = append(table.Rows, group)
			return nil
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TeamGroupTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				teamGroup := row.(entity.TeamGroup)
				if teamGroup.TeamID == group.TeamID &&
					teamGroup.Label == group.Label {
					continue
				}

				rows = append(rows, row)
			}

			table.Rows = rows
			return nil
		},
	})
}

func (t TeamGroup) DeleteGroup(
	ct context.Context,
	tx *transaction.Transaction,
	teamID uint64,
	groupLabel string,
) *errs.Error {
	oldGroup, internalErr := t.FindGroupByTeamIDAndLabelWithTx(ct, tx, teamID, groupLabel)
	if internalErr != nil {
		return internalErr
	}

	return tx.ExecuteCommand(transaction.Command{
		Execute: func() *errs.Error {
			table, err := t.db.GetTable(TeamGroupTableName)
			if err != nil {
				return err
			}

			rows := make([]interface{}, 0)
			for _, row := range table.Rows {
				currGroup := row.(entity.TeamGroup)
				if currGroup.TeamID == teamID ||
					currGroup.Label == groupLabel {
					rows = append(rows, currGroup)
				}
			}

			table.Rows = rows
			return nil
		},
		Undo: func() *errs.Error {
			table, err := t.db.GetTable(TeamGroupTableName)
			if err != nil {
				return err
			}

			table.Rows = append(table.Rows, oldGroup)
			return nil
		},
	})
}

func NewTeamGroup(
	db *dbtest.InMemoryDB,
	transactionFactory transaction.Factory,
) TeamGroup {
	return TeamGroup{
		db:                 db,
		transactionFactory: transactionFactory,
	}
}
