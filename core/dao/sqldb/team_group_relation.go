package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamGroupRelation struct {
	transactionFactory transaction.Factory
}

var _ dao.TeamGroupRelation = (*TeamGroupRelation)(nil)

func (*TeamGroupRelation) FindTeamIDsByGroupIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	groupID uint64,
) ([]uint64, *errs.Error) {
	teamIDs := []uint64{}
	row, err := tx.SQLTx().QueryContext(
		ct,
		`
		SELECT team_id
		FROM team_group_relation
		WHERE group_id = $1;`,
		groupID,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer row.Close()

	for row.Next() {
		var teamID uint64
		err := row.Scan(&teamID)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		teamIDs = append(teamIDs, teamID)
	}

	return teamIDs, nil
}

func (t *TeamGroupRelation) FindTeamIDsByGroupID(ct context.Context, groupID uint64) ([]uint64, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := t.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return t.FindTeamIDsByGroupIDWithTx(ct, tx, groupID)
}

func (*TeamGroupRelation) CreateTeamGroupRelation(
	ct context.Context,
	tx *transaction.Transaction,
	teamGroupRelation entity.TeamGroupRelation,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
		INSERT INTO team_group_relation (
			team_id,
			group_id
		) 
		VALUES ($1, $2);`,
		teamGroupRelation.TeamID,
		teamGroupRelation.GroupID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (*TeamGroupRelation) DeleteTeamGroupRelation(ct context.Context, tx *transaction.Transaction, teamID uint64, groupID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
		DELETE FROM team_group_relation
		WHERE team_id = $1 AND group_id = $2;`,
		teamID,
		groupID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTeamGroupRelation(
	transactionFactory transaction.Factory,
) *TeamGroupRelation {
	return &TeamGroupRelation{
		transactionFactory: transactionFactory,
	}
}
