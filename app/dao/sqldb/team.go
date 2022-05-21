package sqldb

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/teamyapp/teamy-backend/app/dao"
	"github.com/teamyapp/teamy-backend/app/entityv2"
)

type Team struct {
	db *sql.DB
}

var _ dao.Team = (*Team)(nil)

func (t Team) FindAllTeams() ([]entityv2.Team, error) {
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
		log.Println(err)
		return nil, err
	}
	defer rows.Close()

	teams := make([]entityv2.Team, 0)
	for rows.Next() {
		team := entityv2.Team{}
		err = rows.Scan(
			&team.ID,
			&team.Name,
			&team.IconURL,
			&team.CreatorID,
			&team.OwnerID,
			&team.CreatedAt,
			&team.UpdatedAt,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		teams = append(teams, team)
	}

	return teams, err
}

func (t Team) FindTeamByID(teamID uint64) (entityv2.Team, error) {
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
	team := entityv2.Team{}
	err := t.db.QueryRow(statement, teamID).
		Scan(
			&team.ID,
			&team.Name,
			&team.IconURL,
			&team.CreatorID,
			&team.OwnerID,
			&team.CreatedAt,
			&team.UpdatedAt,
		)
	if err != nil {
		log.Println(err)
	}

	return team, err
}

func (t Team) FindTeamsByIDs(teamIDs []uint64) ([]entityv2.Team, error) {
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
	WHERE id IN (%s)`, idsString)
	rows, err := t.db.Query(query)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()
	var teams []entityv2.Team
	for rows.Next() {
		var team entityv2.Team
		err = rows.
			Scan(
				&team.ID,
				&team.Name,
				&team.IconURL,
				&team.CreatorID,
				&team.OwnerID,
				&team.CreatedAt,
				&team.UpdatedAt,
			)
		if err != nil {
			log.Println(team.ID, err)
			continue
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func NewTeam(sqlDB *sql.DB) Team {
	return Team{db: sqlDB}
}
