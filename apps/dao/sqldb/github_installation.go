package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubAppInstallation struct {
	db *sql.DB
}

var _ dao.GithubAppInstallation = (*GithubAppInstallation)(nil)

func (g GithubAppInstallation) CreateGithubAppInstallation(
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
		log.Println(err)
	}

	return err
}

func (g GithubAppInstallation) FindInstallationByID(installationID uint64) (entity.GithubAppInstallation, error) {
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

	return installation, err
}

func NewGithubAppInstallation(sqlDB *sql.DB) GithubAppInstallation {
	return GithubAppInstallation{
		db: sqlDB,
	}
}
