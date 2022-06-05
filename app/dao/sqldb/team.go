package sqldb

import (
	"database/sql"
	"fmt"
	"log"

	"github.com/teamyapp/teamy-backend/app/dao"
	"github.com/teamyapp/teamy-backend/app/entity"
)

type Team struct {
	db *sql.DB
}

var _ dao.Team = (*Team)(nil)

func (t Team) FindAllTeams() ([]entity.Team, error) {
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
			log.Println(err)
			continue
		}

		teams = append(teams, team)
	}

	return teams, err
}

func (t Team) FindTeamByID(teamID uint64) (entity.Team, error) {
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
		log.Println(err)
	}

	return team, err
}

func (t Team) FindTeamsByIDs(teamIDs []uint64) ([]entity.Team, error) {
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
		log.Println(err)
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
			log.Println(team.ID, err)
			continue
		}

		teams = append(teams, team)
	}

	return teams, nil
}

func (t Team) CreateTeam(team entity.Team) error {
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
	return err
}

func (t Team) UpdateTeam(team entity.Team) error {
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
	return err
}

func NewTeam(sqlDB *sql.DB) Team {
	return Team{db: sqlDB}
}
