package sqldb

import (
	"database/sql"
	"errors"
	"fmt"
	"log"

	"github.com/teamyapp/teamy-backend/apps/dao"
	"github.com/teamyapp/teamy-backend/apps/entity"
)

type GithubAppInstallState struct {
	db *sql.DB
}

var _ dao.GithubAppInstallState = (*GithubAppInstallState)(nil)

func (g GithubAppInstallState) FindStateByID(stateID uint64) (entity.GithubAppInstallState, error) {
	state := entity.GithubAppInstallState{}
	err := g.db.QueryRow(`
	SELECT
	    id,
	    team_id,
	    redirect_url,
	    created_at
	FROM apps_github_app_install_state
	WHERE id = $1;
`,
		stateID).
		Scan(
			&state.ID,
			&state.TeamID,
			&state.RedirectURL,
			&state.CreatedAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		return entity.GithubAppInstallState{}, dao.ErrNotFound(fmt.Sprintf(
			"GithubAppInstallState not found: id=%v", stateID))
	}

	return state, err
}

func (g GithubAppInstallState) CreateState(state entity.GithubAppInstallState) error {
	_, err := g.db.Exec(`
	INSERT INTO apps_github_app_install_state
	(
	    id,
	    team_id,
	    redirect_url,
	    created_at
	)
	VALUES ($1, $2, $3, $4);
`,
		int64(state.ID),
		state.TeamID,
		state.RedirectURL,
		state.CreatedAt,
	)
	if err != nil {
		log.Println(err)
	}

	return err
}

func (g GithubAppInstallState) DeleteState(stateID uint64) error {
	_, err := g.db.Exec(`
		DELETE FROM apps_github_app_install_state
		WHERE id = $1;
		`,
		stateID)
	return err
}

func NewGithubAppInstallState(sqlDB *sql.DB) GithubAppInstallState {
	return GithubAppInstallState{
		db: sqlDB,
	}
}
