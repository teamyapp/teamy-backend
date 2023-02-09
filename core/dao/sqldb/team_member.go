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

type TeamMember struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.TeamMember = (*TeamMember)(nil)

func (t TeamMember) FindTeamIDsByUserID(ct context.Context, userID uint64) ([]uint64, *errs.Error) {
	statement := `
	SELECT
		team_id
	FROM team_member
	WHERE user_id = $1;
`
	rows, err := t.db.Query(statement, int64(userID))
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	teamIDs := make([]uint64, 0)
	for rows.Next() {
		var teamID uint64
		err = rows.Scan(
			&teamID,
		)
		if err != nil {
			internalErr = &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}
			t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{
				telemetry.CauseProp: internalErr,
			})
			continue
		}

		teamIDs = append(teamIDs, teamID)
	}

	if internalErr != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	return teamIDs, nil
}

func (t TeamMember) FindTeamMemberIDsByTeamID(ct context.Context, teamID uint64) ([]uint64, *errs.Error) {
	statement := `
	SELECT
		user_id
	FROM team_member
	WHERE team_id = $1;
`
	rows, err := t.db.Query(statement, int64(teamID))
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	teamMemberIDs := make([]uint64, 0)
	for rows.Next() {
		var teamMemberID uint64
		err = rows.Scan(
			&teamMemberID,
		)
		if err != nil {
			internalErr = &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}
			t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
			continue
		}

		teamMemberIDs = append(teamMemberIDs, teamMemberID)
	}

	if internalErr != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	return teamMemberIDs, nil
}

func (t TeamMember) FindTeamMembersByTeamID(ct context.Context, teamID uint64) ([]entity.TeamMember, *errs.Error) {
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	defer rows.Close()
	var internalErr *errs.Error
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
			internalErr = &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}
			t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{
				telemetry.CauseProp: internalErr,
			})
			continue
		}

		teamMembers = append(teamMembers, teamMember)
	}

	if internalErr != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	return teamMembers, nil
}

func (t TeamMember) FindTeamMember(ct context.Context, teamID uint64, userID uint64) (entity.TeamMember, *errs.Error) {
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
		internalErr := &errs.Error{
			Code: errs.NotFound,
			Message: fmt.Sprintf(
				"team member not found: teamID=%v, userID=%v", teamID, userID),
		}
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.TeamMember{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.TeamMember{}, internalErr
	}

	return teamMember, nil
}

func (t TeamMember) CreateTeamMember(ct context.Context, teamMember entity.TeamMember) *errs.Error {
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func (t TeamMember) UpdateTeamMember(ct context.Context, teamMember entity.TeamMember) *errs.Error {
	_, err := t.db.Exec(`
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func (t TeamMember) DeleteTeamMember(ct context.Context, teamID uint64, userID uint64) *errs.Error {
	_, err := t.db.Exec(`
		DELETE FROM team_member
		WHERE team_id = $1 AND user_id = $2;
		`,
		teamID, userID)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func NewTeamMember(dataCollector telemetry.DataCollector, sqlDB *sql.DB) TeamMember {
	return TeamMember{dataCollector: dataCollector, db: sqlDB}
}
