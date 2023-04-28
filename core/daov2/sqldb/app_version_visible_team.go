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

var _ daov2.AppVersionVisibleTeam = (*AppVersionVisibleTeam)(nil)

type AppVersionVisibleTeam struct {
	logger telemetry.Logger
}

func (a AppVersionVisibleTeam) FindAppVersionVisibleTeamWithTx(ct context.Context, tx *transaction.Transaction,
	appID uint64, versionNumber int32, teamID uint64) (entity.AppVersionVisibleTeam, *errs.Error) {
	appVersionVisibleTeam := entity.AppVersionVisibleTeam{}
	err := tx.SQLTx().QueryRow(`
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

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.AppVersionVisibleTeam{}, errs.NewError(errs.NotFound,
				fmt.Sprintf("app version visible team not found: appID=%v, versionNum=%v, teamID=%v", appID, versionNumber, teamID))
		}

		return entity.AppVersionVisibleTeam{}, errs.NewError(errs.Unknown, err.Error())
	}

	return appVersionVisibleTeam, nil
}

func (a AppVersionVisibleTeam) FindAppVersionVisibleTeamsByAppIDAndVersionNumberWithTx(ct context.Context,
	tx *transaction.Transaction, appID uint64, versionNumber int32) ([]entity.AppVersionVisibleTeam, *errs.Error) {
	rows, err := tx.SQLTx().Query(`
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
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	appVersionVisibleTeams := make([]entity.AppVersionVisibleTeam, 0)
	for rows.Next() {
		appVersionVisibleTeam := entity.AppVersionVisibleTeam{}
		err = rows.Scan(
			&appVersionVisibleTeam.AppID,
			&appVersionVisibleTeam.VersionNumber,
			&appVersionVisibleTeam.TeamID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		appVersionVisibleTeams = append(appVersionVisibleTeams, appVersionVisibleTeam)
	}

	return appVersionVisibleTeams, nil
}

func (a AppVersionVisibleTeam) FindAppVersionVisibleTeamsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction,
	teamID uint64) ([]entity.AppVersionVisibleTeam, *errs.Error) {
	rows, err := tx.SQLTx().Query(`
	SELECT
	    app_id,
	    version_number,
	    team_id
	FROM app_version_visible_team
	WHERE team_id = $1;
`,
		teamID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	appVersionVisibleTeams := make([]entity.AppVersionVisibleTeam, 0)
	for rows.Next() {
		appVersionVisibleTeam := entity.AppVersionVisibleTeam{}
		err = rows.Scan(
			&appVersionVisibleTeam.AppID,
			&appVersionVisibleTeam.VersionNumber,
			&appVersionVisibleTeam.TeamID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		appVersionVisibleTeams = append(appVersionVisibleTeams, appVersionVisibleTeam)
	}

	return appVersionVisibleTeams, nil
}

func (a AppVersionVisibleTeam) CreateAppVersionVisibleTeam(ct context.Context, tx *transaction.Transaction,
	appVersionVisibleTeam entity.AppVersionVisibleTeam) *errs.Error {
	_, err := tx.SQLTx().Exec(`
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
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a AppVersionVisibleTeam) DeleteAppVersionVisibleTeam(ct context.Context, tx *transaction.Transaction,
	appID uint64, versionNumber int32, teamID uint64) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM app_version_visible_team
		WHERE app_id = $1
		AND version_number = $2
		AND team_id = $3;
		`,
		appID,
		versionNumber,
		teamID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a AppVersionVisibleTeam) DeleteAppVersionVisibleTeamsByAppIDAndVersionNumber(ct context.Context,
	tx *transaction.Transaction, appID uint64, versionNumber int32) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM app_version_visible_team
		WHERE app_id = $1
		AND version_number = $2;
		`,
		appID,
		versionNumber)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewAppVersionVisibleTeam(logger telemetry.Logger) AppVersionVisibleTeam {
	return AppVersionVisibleTeam{logger: logger}
}
