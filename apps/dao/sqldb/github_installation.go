package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubAppInstallation struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.GithubAppInstallation = (*GithubAppInstallation)(nil)

func (g GithubAppInstallation) CreateGithubAppInstallation(
	ct context.Context,
	installation entity.GithubAppInstallation,
) *errs.Error {
	_, err := g.db.Exec(`
	INSERT INTO apps_github_app_installation
	(
	    id,
	    team_id,
	    created_at
	)
	VALUES ($1, $2, $3);
`,
		installation.ID,
		installation.TeamID,
		installation.CreatedAt,
	)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (g GithubAppInstallation) FindInstallationIDByTeamID(ct context.Context, teamID uint64) (int, *errs.Error) {
	var installationID int
	err := g.db.QueryRow(`
		SELECT id
		FROM apps_github_app_installation
		WHERE team_id = $1;`,
		teamID).
		Scan(&installationID)
	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("installation not found, team_id=%d", teamID),
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return 0, internalErr
	}

	return installationID, nil
}

func (g GithubAppInstallation) FindInstallationByID(
	ct context.Context,
	installationID uint64,
) (entity.GithubAppInstallation, *errs.Error) {
	installation := entity.GithubAppInstallation{}
	err := g.db.QueryRow(`
	SELECT
	    id,
	    team_id,
	    created_at
	FROM apps_github_app_installation
	WHERE id = $1;
`,
		installationID).
		Scan(
			&installation.ID,
			&installation.TeamID,
			&installation.CreatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code: errs.NotFound,
			Message: fmt.Sprintf(
				"GithubAppInstallation not found: id=%v", installationID),
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.GithubAppInstallation{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.GithubAppInstallation{}, internalErr
	}

	return installation, nil
}

func NewGithubAppInstallation(dataCollector telemetry.DataCollector, sqlDB *sql.DB) GithubAppInstallation {
	return GithubAppInstallation{
		dataCollector: dataCollector,
		db:            sqlDB,
	}
}
