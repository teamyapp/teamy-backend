package sqldb

import (
	"database/sql"
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

func NewGithubAppInstallation(sqlDB *sql.DB) GithubAppInstallation {
	return GithubAppInstallation{
		db: sqlDB,
	}
}
