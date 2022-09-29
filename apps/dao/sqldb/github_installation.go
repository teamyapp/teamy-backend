package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/obs"
	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubAppInstallation struct {
	dataCollector obs.DataCollector
	db            *sql.DB
}

var _ dao.GithubAppInstallation = (*GithubAppInstallation)(nil)

func (g GithubAppInstallation) CreateGithubAppInstallation(
	ct context.Context,
	installation entity.GithubAppInstallation,
) error {
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
		g.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return err
}

func (g GithubAppInstallation) FindInstallationByID(ct context.Context, installationID uint64) (entity.GithubAppInstallation, error) {
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
		return entity.GithubAppInstallation{}, dao.ErrNotFound(fmt.Sprintf(
			"GithubAppInstallation not found: id=%v", installationID))
	}

	if err != nil {
		g.dataCollector.Logger.LogWithContext(ct, obs.Error, obs.Props{obs.CauseProp: err})
	}

	return installation, err
}

func NewGithubAppInstallation(dataCollector obs.DataCollector, sqlDB *sql.DB) GithubAppInstallation {
	return GithubAppInstallation{
		dataCollector: dataCollector,
		db:            sqlDB,
	}
}
