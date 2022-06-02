package sqldb

import (
	"database/sql"
	"log"
	"time"

	"github.com/teamyapp/teamy-backend/app/dao"
)

type TeamMember struct {
	db *sql.DB
}

var _ dao.TeamMember = (*TeamMember)(nil)

func (t TeamMember) FindTeamIDsByUserID(userID uint64) ([]uint64, error) {
	statement := `
	SELECT
		team_id
	FROM team_member
	WHERE user_id = $1;
`
	rows, err := t.db.Query(statement, int64(userID))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()
	teamIDs := make([]uint64, 0)
	for rows.Next() {
		var teamID uint64
		err = rows.Scan(
			&teamID,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		teamIDs = append(teamIDs, teamID)
	}

	return teamIDs, err
}

func (t TeamMember) FindTeamMemberIDsByTeamID(teamID uint64) ([]uint64, error) {
	statement := `
	SELECT
		user_id
	FROM team_member
	WHERE team_id = $1;
`
	rows, err := t.db.Query(statement, int64(teamID))
	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()
	teamMemberIDs := make([]uint64, 0)
	for rows.Next() {
		var teamMemberID uint64
		err = rows.Scan(
			&teamMemberID,
		)
		if err != nil {
			log.Println(err)
			continue
		}

		teamMemberIDs = append(teamMemberIDs, teamMemberID)
	}

	return teamMemberIDs, err
}

func (t TeamMember) HasTeamMember(teamID uint64, userID uint64) (bool, error) {
	statement := `
	SELECT
		*
	FROM team_member
	WHERE team_id = $1 AND user_id = $2;
`
	rows, err := t.db.Query(statement, int64(teamID), int64(userID))
	if err != nil {
		return false, err
	}

	return rows.Next(), nil
}

func (t TeamMember) CreateTeamMember(teamID uint64, userID uint64) error {
	_, err := t.db.Exec(`
		INSERT INTO team_member
		(
		 	team_id,
		 	user_id,
		 	created_at
		)
		VALUES ($1, $2, $3);`,
		teamID,
		userID,
		time.Now(),
	)
	return err
}

func (t TeamMember) DeleteTeamMember(teamID uint64, userID uint64) error {
	_, err := t.db.Exec(`
		DELETE FROM team_member
		WHERE team_id = $1 AND user_id = $2;
		`,
		teamID, userID)
	return err
}

func NewTeamMember(sqlDB *sql.DB) TeamMember {
	return TeamMember{db: sqlDB}
}
