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

type GithubAppInstallState struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.GithubAppInstallState = (*GithubAppInstallState)(nil)

func (g GithubAppInstallState) FindStateByID(
	ct context.Context,
	stateID uint64,
) (entity.GithubAppInstallState, *errs.Error) {
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
		internalErr := &errs.Error{
			Code: errs.NotFound,
			Message: fmt.Sprintf(
				"GithubAppInstallState not found: stateID=%v", stateID),
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.GithubAppInstallState{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return entity.GithubAppInstallState{}, internalErr
	}

	return state, nil
}

func (g GithubAppInstallState) CreateState(
	ct context.Context,
	state entity.GithubAppInstallState,
) *errs.Error {
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
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		g.dataCollector.Logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (g GithubAppInstallState) DeleteState(ct context.Context, stateID uint64) *errs.Error {
	_, err := g.db.Exec(`
		DELETE FROM apps_github_app_install_state
		WHERE id = $1;
		`,
		stateID)

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

func NewGithubAppInstallState(dataCollector telemetry.DataCollector, sqlDB *sql.DB) GithubAppInstallState {
	return GithubAppInstallState{
		dataCollector: dataCollector,
		db:            sqlDB,
	}
}
