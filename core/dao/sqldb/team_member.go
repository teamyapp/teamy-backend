package sqldb

import (
	"context"
	"database/sql"
	"time"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
)

type TeamMember struct {
	dataCollector obs.DataCollector
	db            *sql.DB
}

var _ dao.TeamMember = (*TeamMember)(nil)

func (t TeamMember) FindTeamIDsByUserID(ct context.Context, userID uint64) ([]uint64, error) {
	statement := `
	SELECT
		team_id
	FROM team_member
	WHERE user_id = $1;
`
	rows, err := t.db.Query(statement, int64(userID))
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		teamIDs = append(teamIDs, teamID)
	}

	return teamIDs, nil
}

func (t TeamMember) FindTeamMemberIDsByTeamID(ct context.Context, teamID uint64) ([]uint64, error) {
	statement := `
	SELECT
		user_id
	FROM team_member
	WHERE team_id = $1;
`
	rows, err := t.db.Query(statement, int64(teamID))
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
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
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		teamMemberIDs = append(teamMemberIDs, teamMemberID)
	}

	return teamMemberIDs, err
}

func (t TeamMember) HasTeamMember(ct context.Context, teamID uint64, userID uint64) (bool, error) {
	statement := `
	SELECT
		*
	FROM team_member
	WHERE team_id = $1 AND user_id = $2;
`
	rows, err := t.db.Query(statement, int64(teamID), int64(userID))
	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return false, err
	}

	return rows.Next(), nil
}

func (t TeamMember) CreateTeamMember(ct context.Context, teamID uint64, userID uint64) error {
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

	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (t TeamMember) DeleteTeamMember(ct context.Context, teamID uint64, userID uint64) error {
	_, err := t.db.Exec(`
		DELETE FROM team_member
		WHERE team_id = $1 AND user_id = $2;
		`,
		teamID, userID)

	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func NewTeamMember(dataCollector obs.DataCollector, sqlDB *sql.DB) TeamMember {
	return TeamMember{dataCollector: dataCollector, db: sqlDB}
}
