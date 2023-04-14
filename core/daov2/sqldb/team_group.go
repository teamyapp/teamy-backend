package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/daov2"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamGroup struct {
	logger             telemetry.Logger
	transactionFactory transaction.Factory
}

func (t TeamGroup) FindGroupByTeamIDAndLabel(
	ct context.Context,
	teamID uint64,
	groupLabel string,
) (entity.TeamGroup, *errs.Error) {
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
	statement := `
	SELECT
		team_id,
		label,
		user_group_id,
		created_at
	FROM team_group
	WHERE team_id = $1 AND label = $2;
`
	teamGroup := entity.TeamGroup{}
	err := tx.SQLTx().QueryRow(statement, teamID, groupLabel).
		Scan(
			&teamGroup.TeamID,
			&teamGroup.Label,
			&teamGroup.UserGroupID,
			&teamGroup.CreatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.TeamGroup{}, errs.NewError(
			errs.NotFound,
			fmt.Sprintf("team group not found: team_id=%v, label=%v",
				teamID,
				groupLabel))
	}

	if err != nil {
		return entity.TeamGroup{}, errs.NewError(errs.Unknown, err.Error())
	}

	return teamGroup, nil
}

func (t TeamGroup) CreateGroup(
	ct context.Context,
	tx *transaction.Transaction,
	group entity.TeamGroup,
) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		INSERT INTO team_group
		(
			team_id,
			label,
			user_group_id,
			created_at
		)
		VALUES ($1, $2, $3, $4);`,
		group.TeamID,
		group.Label,
		group.UserGroupID,
		group.CreatedAt,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t TeamGroup) DeleteGroup(
	ct context.Context,
	tx *transaction.Transaction,
	teamID uint64,
	groupLabel string,
) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM team_group
		WHERE team_id = $1 AND label = $2;`,
		teamID,
		groupLabel,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

var _ daov2.TeamGroup = (*TeamGroup)(nil)

func NewTeamGroup(logger telemetry.Logger, transactionFactory transaction.Factory) TeamGroup {
	return TeamGroup{logger: logger, transactionFactory: transactionFactory}
}
