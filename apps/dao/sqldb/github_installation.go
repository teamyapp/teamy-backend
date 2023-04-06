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
	logger telemetry.Logger
	db     *sql.DB
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
		return errs.NewError(errs.Unknown, err.Error())
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

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, errs.NewError(errs.NotFound, fmt.Sprintf(
				"installation not found: teamID=%d", teamID))
		}

		return 0, errs.NewError(errs.Unknown, err.Error())
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

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.GithubAppInstallation{}, errs.NewError(errs.NotFound, fmt.Sprintf(
				"GithubAppInstallation not found: id=%v", installationID))
		}

		return entity.GithubAppInstallation{}, errs.NewError(errs.Unknown, err.Error())
	}

	return installation, nil
}

func (g GithubAppInstallation) FindInstallationByTeamID(ct context.Context, teamID uint64) (entity.GithubAppInstallation, *errs.Error) {
	installation := entity.GithubAppInstallation{}
	err := g.db.QueryRow(`
	SELECT
	    id,
	    team_id,
	    created_at
	FROM apps_github_app_installation
	WHERE team_id = $1;
`,
		teamID).
		Scan(
			&installation.ID,
			&installation.TeamID,
			&installation.CreatedAt,
		)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.GithubAppInstallation{}, errs.NewError(errs.NotFound, fmt.Sprintf(
				"GithubAppInstallation not found: teamId=%v", teamID))
		}

		return entity.GithubAppInstallation{}, errs.NewError(errs.Unknown, err.Error())
	}

	return installation, nil
}

func NewGithubAppInstallation(logger telemetry.Logger, sqlDB *sql.DB) GithubAppInstallation {
	return GithubAppInstallation{
		logger: logger,
		db:     sqlDB,
	}
}
