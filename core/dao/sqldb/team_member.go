package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamMember struct {
	logger             telemetry.Logger
	transactionFactory transaction.Factory
}

var _ dao.TeamMember = (*TeamMember)(nil)

func (t TeamMember) FindTeamIDsByUserID(ct context.Context, userID uint64) ([]uint64, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := t.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return t.FindTeamIDsByUserIDWithTx(ct, tx, userID)
}

func (t TeamMember) FindTeamMembersByTeamID(ct context.Context, teamID uint64) ([]entity.TeamMember, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := t.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return t.FindTeamMembersByTeamIDWithTx(ct, tx, teamID)
}

func (t TeamMember) FindTeamIDsByUserIDWithTx(ct context.Context, tx *transaction.Transaction, userID uint64) ([]uint64, *errs.Error) {
	statement := `
	SELECT
		team_id
	FROM team_member
	WHERE user_id = $1;
`
	rows, err := tx.SQLTx().Query(statement, int64(userID))
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	teamIDs := make([]uint64, 0)
	for rows.Next() {
		var teamID uint64
		err = rows.Scan(
			&teamID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		teamIDs = append(teamIDs, teamID)
	}

	return teamIDs, nil
}

func (t TeamMember) FindTeamMemberIDsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]uint64, *errs.Error) {
	statement := `
	SELECT
		user_id
	FROM team_member
	WHERE team_id = $1;
`
	rows, err := tx.SQLTx().Query(statement, int64(teamID))
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	teamMemberIDs := make([]uint64, 0)
	for rows.Next() {
		var teamMemberID uint64
		err = rows.Scan(
			&teamMemberID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		teamMemberIDs = append(teamMemberIDs, teamMemberID)
	}

	return teamMemberIDs, nil
}

func (t TeamMember) FindTeamMembersByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.TeamMember, *errs.Error) {
	rows, err := tx.SQLTx().Query(`
	SELECT
		team_id,
		user_id,
		weekly_bandwidth,
		created_at,
		updated_at
	FROM team_member
	WHERE team_id = $1;
`, teamID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	teamMembers := make([]entity.TeamMember, 0)
	for rows.Next() {
		var teamMember entity.TeamMember
		err = rows.Scan(
			&teamMember.TeamID,
			&teamMember.UserID,
			&teamMember.WeeklyBandwidth,
			&teamMember.CreatedAt,
			&teamMember.UpdatedAt,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		teamMembers = append(teamMembers, teamMember)
	}

	return teamMembers, nil
}

func (t TeamMember) FindTeamMemberWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64, userID uint64) (entity.TeamMember, *errs.Error) {
	teamMember := entity.TeamMember{}
	err := tx.SQLTx().QueryRow(
		`
	SELECT
		team_id,
		user_id,
		weekly_bandwidth,
		created_at,
		updated_at
	FROM team_member
	WHERE team_id = $1 AND user_id=$2;
`,
		teamID,
		userID).
		Scan(
			&teamMember.TeamID,
			&teamMember.UserID,
			&teamMember.WeeklyBandwidth,
			&teamMember.CreatedAt,
			&teamMember.UpdatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.TeamMember{}, errs.NewError(errs.NotFound, fmt.Sprintf(
			"team member not found: teamID=%v, userID=%v", teamID, userID))
	}

	if err != nil {
		return entity.TeamMember{}, errs.NewError(errs.Unknown, err.Error())
	}

	return teamMember, nil
}

func (t TeamMember) CreateTeamMember(ct context.Context, tx *transaction.Transaction, teamMember entity.TeamMember) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		INSERT INTO team_member
		(
		 	team_id,
			user_id,
			weekly_bandwidth,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5);`,
		teamMember.TeamID,
		teamMember.UserID,
		teamMember.WeeklyBandwidth,
		teamMember.CreatedAt,
		teamMember.UpdatedAt,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t TeamMember) UpdateTeamMember(ct context.Context, tx *transaction.Transaction, teamMember entity.TeamMember) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		UPDATE team_member
		SET
			weekly_bandwidth = $1,
			updated_at = $2
		WHERE team_id = $3 AND user_id = $4;`,
		teamMember.WeeklyBandwidth,
		teamMember.UpdatedAt,
		teamMember.TeamID,
		teamMember.UserID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t TeamMember) DeleteTeamMember(ct context.Context, tx *transaction.Transaction, teamID uint64, userID uint64) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM team_member
		WHERE team_id = $1 AND user_id = $2;
		`,
		teamID, userID)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTeamMember(logger telemetry.Logger, transactionFactory transaction.Factory) TeamMember {
	return TeamMember{logger: logger, transactionFactory: transactionFactory}
}
