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

type Team struct {
	dataCollector      telemetry.DataCollector
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
	statement := `
	SELECT
		id,
		name,
		icon_url,
		creator_id,
		owner_id,
		created_at,
		updated_at
	FROM team;
`
	rows, err := tx.SQLTx().Query(statement)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	teams := make([]entity.Team, 0)
	for rows.Next() {
		team := entity.Team{}
		err = rows.Scan(
			&team.ID,
			&team.Name,
			&team.IconURL,
			&team.CreatorUserID,
			&team.OwnerUserID,
			&team.CreatedAt,
			&team.UpdatedAt,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func (t Team) FindTeamByIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) (entity.Team, *errs.Error) {
	statement := `
	SELECT
		id,
		name,
		icon_url,
		creator_id,
		owner_id,
		created_at,
		updated_at
	FROM team
	WHERE id = $1;
`
	team := entity.Team{}
	err := tx.SQLTx().QueryRow(statement, teamID).
		Scan(
			&team.ID,
			&team.Name,
			&team.IconURL,
			&team.CreatorUserID,
			&team.OwnerUserID,
			&team.CreatedAt,
			&team.UpdatedAt,
		)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.Team{}, errs.NewError(errs.NotFound, fmt.Sprintf("team not found: teamID=%v", teamID))
		}
		
		return entity.Team{}, errs.NewError(errs.Unknown, err.Error())
	}

	if err != nil {
		return entity.Team{}, errs.NewError(errs.Unknown, err.Error())
	}

	return team, nil
}

func (t Team) FindTeamsByIDsWithTx(ct context.Context, tx *transaction.Transaction, teamIDs []uint64) ([]entity.Team, *errs.Error) {
	if len(teamIDs) == 0 {
		return []entity.Team{}, nil
	}

	idsString := toIDsString(teamIDs)
	query := fmt.Sprintf(`
	SELECT
		id,
		name,
		icon_url,
		creator_id,
		owner_id,
		created_at,
		updated_at
	FROM team
	WHERE id IN (%v);`, idsString)
	rows, err := tx.SQLTx().Query(query)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	var teams []entity.Team
	for rows.Next() {
		var team entity.Team
		err = rows.
			Scan(
				&team.ID,
				&team.Name,
				&team.IconURL,
				&team.CreatorUserID,
				&team.OwnerUserID,
				&team.CreatedAt,
				&team.UpdatedAt,
			)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func (t Team) CreateTeam(ct context.Context, tx *transaction.Transaction, team entity.Team) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		INSERT INTO team
		    (
				 id,
				 name,
				 creator_id,
				 owner_id,
				 created_at
		    )
		VALUES ($1, $2, $3, $4, $5);`,
		team.ID,
		team.Name,
		team.CreatorUserID,
		team.OwnerUserID,
		team.CreatedAt,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t Team) UpdateTeam(ct context.Context, tx *transaction.Transaction, team entity.Team) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		UPDATE team
		SET
			name = $1,
			icon_url = $2,
			owner_id = $3,
			updated_at = $4
		WHERE id = $5;`,
		team.Name,
		team.IconURL,
		team.OwnerUserID,
		team.UpdatedAt,
		team.ID,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t Team) DeleteTeam(ct context.Context, tx *transaction.Transaction, teamID uint64) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM team
		WHERE id = $1;`,
		teamID,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTeam(dataCollector telemetry.DataCollector, transactionFactory transaction.Factory) Team {
	return Team{dataCollector: dataCollector, transactionFactory: transactionFactory}
}
