package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
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

func (t TeamMember) FindTeamMembersByTeamID(ct context.Context, teamID uint64) ([]entity.TeamMember, error) {
	rows, err := t.db.Query(`
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
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
		return nil, err
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
			t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
			continue
		}

		teamMembers = append(teamMembers, teamMember)
	}

	return teamMembers, err
}

func (t TeamMember) FindTeamMember(ct context.Context, teamID uint64, userID uint64) (entity.TeamMember, error) {
	teamMember := entity.TeamMember{}
	err := t.db.QueryRow(
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
		return entity.TeamMember{}, dao.ErrNotFound(fmt.Sprintf(
			"team member not found: teamID=%v, userID=%v",
			teamMember.TeamID,
			teamMember.UserID))
	}

	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return teamMember, err
}

func (t TeamMember) CreateTeamMember(ct context.Context, teamMember entity.TeamMember) error {
	_, err := t.db.Exec(`
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
		t.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (t TeamMember) UpdateTeamMember(ct context.Context, teamMember entity.TeamMember) error {
	_, err := t.db.Exec(`
		UPDATE team_member
		SET
			weekly_bandwidth = $1,
			created_at = $2,
			updated_at = $3
		WHERE team_id = $4 AND user_id = $5;`,
		teamMember.WeeklyBandwidth,
		teamMember.CreatedAt,
		teamMember.UpdatedAt,
		teamMember.TeamID,
		teamMember.UserID,
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
