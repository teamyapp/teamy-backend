package sqldb

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type Team struct {
	dataCollector obs.DataCollector
	db            *sql.DB
}

var _ dao.Team = (*Team)(nil)

func (t Team) FindAllTeams(ct context.Context) ([]entity.Team, error) {
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
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
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
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func (t Team) FindTeamByID(ct context.Context, teamID uint64) (entity.Team, error) {
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
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return team, err
}

func (t Team) FindTeamsByIDs(ct context.Context, teamIDs []uint64) ([]entity.Team, error) {
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
	WHERE id IN (%s);`, idsString)
	rows, err := t.db.Query(query)
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
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
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func (t Team) CreateTeam(ct context.Context, team entity.Team) error {
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
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (t Team) UpdateTeam(ct context.Context, team entity.Team) error {
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
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewTeam(dataCollector obs.DataCollector, sqlDB *sql.DB) Team {
	return Team{dataCollector: dataCollector, db: sqlDB}
}
