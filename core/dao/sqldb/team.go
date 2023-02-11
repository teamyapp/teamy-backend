package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Team struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.Team = (*Team)(nil)

func (t Team) FindAllTeams(ct context.Context) ([]entity.Team, *errs.Error) {
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
	rows, err := t.db.Query(statement)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
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
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			t.dataCollector.Logger.ErrorWithContext(ct, newInternalErr)
			continue
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func (t Team) FindTeamByID(ct context.Context, teamID uint64) (entity.Team, *errs.Error) {
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
	err := t.db.QueryRow(statement, teamID).
		Scan(
			&team.ID,
			&team.Name,
			&team.IconURL,
			&team.CreatorUserID,
			&team.OwnerUserID,
			&team.CreatedAt,
			&team.UpdatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code: errs.NotFound,
			Message: fmt.Sprintf(
				"team not found: teamID=%v", teamID),
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Team{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.Team{}, internalErr
	}

	return team, nil
}

func (t Team) FindTeamsByIDs(ct context.Context, teamIDs []uint64) ([]entity.Team, *errs.Error) {
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
	rows, err := t.db.Query(query)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
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
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			t.dataCollector.Logger.ErrorWithContext(ct, newInternalErr)
			continue
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func (t Team) CreateTeam(ct context.Context, team entity.Team) *errs.Error {
	_, err := t.db.Exec(`
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (t Team) UpdateTeam(ct context.Context, team entity.Team) *errs.Error {
	_, err := t.db.Exec(`
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (t Team) DeleteTeam(ct context.Context, teamID uint64) *errs.Error {
	_, err := t.db.Exec(`
		DELETE FROM team
		WHERE id = $1;`,
		teamID,
	)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func NewTeam(dataCollector telemetry.DataCollector, sqlDB *sql.DB) Team {
	return Team{dataCollector: dataCollector, db: sqlDB}
}
