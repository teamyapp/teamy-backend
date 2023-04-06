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

type TeamFileUploadSession struct {
	logger telemetry.Logger
}

var _ daov2.TeamFileUploadSession = (*TeamFileUploadSession)(nil)

func (t TeamFileUploadSession) FindTeamFileUploadSessionByTeamIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	teamID uint64,
	teamFileUploadSessionType entity.TeamFileUploadSessionType,
	fileUploadSessionID uint64,
) (entity.TeamFileUploadSession, *errs.Error) {
	teamFileUploadSession := entity.TeamFileUploadSession{}
	err := tx.SQLTx().QueryRow(`
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

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.TeamFileUploadSession{}, errs.NewError(errs.NotFound, fmt.Sprintf(
				"TeamFileUploadSession not found: teamID=%v, teamFileUploadSessionType=%v", teamID, teamFileUploadSessionType))
		}

		return entity.TeamFileUploadSession{}, errs.NewError(errs.Unknown, err.Error())
	}

	return teamFileUploadSession, nil
}

func (t TeamFileUploadSession) CreateTeamFileUploadSession(
	ct context.Context,
	tx *transaction.Transaction,
	teamFileUploadSession entity.TeamFileUploadSession,
) *errs.Error {
	_, err := tx.SQLTx().Exec(`
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
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t TeamFileUploadSession) UpdateTeamFileUploadSession(
	ct context.Context,
	tx *transaction.Transaction,
	teamFileUploadSession entity.TeamFileUploadSession,
) *errs.Error {
	_, err := tx.SQLTx().Exec(`
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
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTeamFileUploadSession(logger telemetry.Logger) TeamFileUploadSession {
	return TeamFileUploadSession{logger: logger}
}
