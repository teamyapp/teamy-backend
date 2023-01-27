package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamFileUploadSession struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.TeamFileUploadSession = (*TeamFileUploadSession)(nil)

func (t TeamFileUploadSession) FindTeamFileUploadSessionByTeamID(
	ct context.Context,
	teamID uint64,
	teamFileUploadSessionType entity.TeamFileUploadSessionType,
	fileUploadSessionID uint64,
) (entity.TeamFileUploadSession, error) {
	teamFileUploadSession := entity.TeamFileUploadSession{}
	err := t.db.QueryRow(`
		SELECT
			team_id,
			type,
			file_upload_session_id,
			is_completed,
			created_at,
			updated_at
		FROM team_file_upload_session
		WHERE team_id = $1 AND type = $2 AND file_upload_session_id = $3;`,
		teamID, teamFileUploadSessionType, fileUploadSessionID).
		Scan(
			&teamFileUploadSession.TeamID,
			&teamFileUploadSession.Type,
			&teamFileUploadSession.FileUploadSessionID,
			&teamFileUploadSession.IsCompleted,
			&teamFileUploadSession.CreatedAt,
			&teamFileUploadSession.UpdatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.TeamFileUploadSession{}, dao.ErrNotFound(fmt.Sprintf(
			"TeamFileUploadSession not found: teamID=%v, type=%v",
			teamID,
			teamFileUploadSessionType))
	}

	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return teamFileUploadSession, err
}

func (t TeamFileUploadSession) CreateTeamFileUploadSession(
	ct context.Context,
	teamFileUploadSession entity.TeamFileUploadSession,
) error {
	_, err := t.db.Exec(`
		INSERT INTO team_file_upload_session
		(
			team_id,
			type,
			file_upload_session_id,
			is_completed,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6);`,
		teamFileUploadSession.TeamID,
		teamFileUploadSession.Type,
		teamFileUploadSession.FileUploadSessionID,
		teamFileUploadSession.IsCompleted,
		teamFileUploadSession.CreatedAt,
		teamFileUploadSession.UpdatedAt,
	)

	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func (t TeamFileUploadSession) UpdateTeamFileUploadSession(
	ct context.Context,
	teamFileUploadSession entity.TeamFileUploadSession,
) error {
	_, err := t.db.Exec(`
		UPDATE team_file_upload_session
		SET
			team_id = $1,
			type = $2,
			file_upload_session_id = $3,
			is_completed = $4,
			created_at= $5,
			updated_at = $6
		WHERE team_id = $7 AND type = $8 AND file_upload_session_id = $9;`,
		teamFileUploadSession.TeamID,
		teamFileUploadSession.Type,
		teamFileUploadSession.FileUploadSessionID,
		teamFileUploadSession.IsCompleted,
		teamFileUploadSession.CreatedAt,
		teamFileUploadSession.UpdatedAt,
		teamFileUploadSession.TeamID,
		teamFileUploadSession.Type,
		teamFileUploadSession.FileUploadSessionID,
	)

	if err != nil {
		t.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func NewTeamFileUploadSession(dataCollector telemetry.DataCollector, sqlDB *sql.DB) TeamFileUploadSession {
	return TeamFileUploadSession{dataCollector: dataCollector, db: sqlDB}
}
