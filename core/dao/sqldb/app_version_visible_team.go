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

var _ dao.AppVersionVisibleTeam = (*AppVersionVisibleTeam)(nil)

type AppVersionVisibleTeam struct {
	logger telemetry.Logger
	db     *sql.DB
}

func (a AppVersionVisibleTeam) FindAppVersionVisibleTeam(ct context.Context, appID uint64, versionNumber int32, teamID uint64) (entity.AppVersionVisibleTeam, *errs.Error) {
	appVersionVisibleTeam := entity.AppVersionVisibleTeam{}
	err := a.db.QueryRow(`
	SELECT
	    app_id,
	    version_number,
	    team_id
	FROM app_version_visible_team
	WHERE app_id = $1 AND version_number = $2 AND team_id = $3;
`,
		appID,
		versionNumber,
		teamID).
		Scan(
			&appVersionVisibleTeam.AppID,
			&appVersionVisibleTeam.VersionNumber,
			&appVersionVisibleTeam.TeamID,
		)

	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("app version visible team not found: appID=%v, versionNum=%v, teamID=%v", appID, versionNumber, teamID),
		}
		a.logger.ErrorWithContext(ct, internalErr)
		return entity.AppVersionVisibleTeam{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.logger.ErrorWithContext(ct, internalErr)
		return entity.AppVersionVisibleTeam{}, internalErr
	}

	return appVersionVisibleTeam, nil
}

func (a AppVersionVisibleTeam) FindAppVersionVisibleTeamsByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int32) ([]entity.AppVersionVisibleTeam, *errs.Error) {
	rows, err := a.db.Query(`
	SELECT
	    app_id,
	    version_number,
	    team_id
	FROM app_version_visible_team
	WHERE app_id = $1 AND version_number = $2;
`,
		appID,
		versionNumber)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	appVersionVisibleTeams := make([]entity.AppVersionVisibleTeam, 0)
	for rows.Next() {
		appVersionVisibleTeam := entity.AppVersionVisibleTeam{}
		err = rows.Scan(
			&appVersionVisibleTeam.AppID,
			&appVersionVisibleTeam.VersionNumber,
			&appVersionVisibleTeam.TeamID,
		)
		if err != nil {
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			a.logger.ErrorWithContext(ct, newInternalErr)
			continue
		}

		appVersionVisibleTeams = append(appVersionVisibleTeams, appVersionVisibleTeam)
	}

	return appVersionVisibleTeams, nil
}

func (a AppVersionVisibleTeam) FindAppVersionVisibleTeamsByTeamID(ct context.Context, teamID uint64) ([]entity.AppVersionVisibleTeam, *errs.Error) {
	rows, err := a.db.Query(`
	SELECT
	    app_id,
	    version_number,
	    team_id
	FROM app_version_visible_team
	WHERE team_id = $1;
`,
		teamID)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.logger.ErrorWithContext(ct, internalErr)
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	appVersionVisibleTeams := make([]entity.AppVersionVisibleTeam, 0)
	for rows.Next() {
		appVersionVisibleTeam := entity.AppVersionVisibleTeam{}
		err = rows.Scan(
			&appVersionVisibleTeam.AppID,
			&appVersionVisibleTeam.VersionNumber,
			&appVersionVisibleTeam.TeamID,
		)
		if err != nil {
			newInternalErr := &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}

			if internalErr == nil {
				internalErr = newInternalErr
			}

			a.logger.ErrorWithContext(ct, newInternalErr)
			continue
		}

		appVersionVisibleTeams = append(appVersionVisibleTeams, appVersionVisibleTeam)
	}

	return appVersionVisibleTeams, nil
}

func (a AppVersionVisibleTeam) CreateAppVersionVisibleTeam(ct context.Context, appVersionVisibleTeam entity.AppVersionVisibleTeam) *errs.Error {
	_, err := a.db.Exec(`
	INSERT INTO app_version_visible_team
	(
	 	app_id,
	 	version_number,
	 	team_id
	)
	VALUES ($1, $2, $3);
`,
		appVersionVisibleTeam.AppID,
		appVersionVisibleTeam.VersionNumber,
		appVersionVisibleTeam.TeamID,
	)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (a AppVersionVisibleTeam) DeleteAppVersionVisibleTeam(ct context.Context, appID uint64, versionNumber int32, teamID uint64) *errs.Error {
	_, err := a.db.Exec(`
		DELETE FROM app_version_visible_team
		WHERE app_id = $1
		AND version_number = $2
		AND team_id = $3;
		`,
		appID,
		versionNumber,
		teamID)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func (a AppVersionVisibleTeam) DeleteAppVersionVisibleTeamsByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int32) *errs.Error {
	_, err := a.db.Exec(`
		DELETE FROM app_version_visible_team
		WHERE app_id = $1
		AND version_number = $2;
		`,
		appID,
		versionNumber)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.logger.ErrorWithContext(ct, internalErr)
		return internalErr
	}

	return nil
}

func NewAppVersionVisibleTeam(logger telemetry.Logger, sqlDB *sql.DB) AppVersionVisibleTeam {
	return AppVersionVisibleTeam{logger: logger, db: sqlDB}
}
