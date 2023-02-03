package sqldb

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/teamyapp/cloud/libs/telemetry"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

var _ dao.AppVersionVisibleTeam = (*AppVersionVisibleTeam)(nil)

type AppVersionVisibleTeam struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

func (a AppVersionVisibleTeam) FindAppVersionVisibleTeam(ct context.Context, appID uint64, versionNumber int32, teamID uint64) (entity.AppVersionVisibleTeam, error) {
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
		return entity.AppVersionVisibleTeam{}, dao.ErrNotFound(fmt.Sprintf(
			"app version visible team not found: appId=%v, versionNumber=%v, teamId=%v", appID, versionNumber, teamID))
	}

	if err != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return appVersionVisibleTeam, err
}

func (a AppVersionVisibleTeam) FindAppVersionVisibleTeamsByAppIDAndVersionNumber(ct context.Context, appID uint64, versionNumber int32) ([]entity.AppVersionVisibleTeam, error) {
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
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
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
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			continue
		}

		appVersionVisibleTeams = append(appVersionVisibleTeams, appVersionVisibleTeam)
	}

	return appVersionVisibleTeams, err
}

func (a AppVersionVisibleTeam) FindAppVersionVisibleTeamsByTeamID(ct context.Context, teamID uint64) ([]entity.AppVersionVisibleTeam, error) {
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
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
		return nil, err
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
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
			continue
		}

		appVersionVisibleTeams = append(appVersionVisibleTeams, appVersionVisibleTeam)
	}

	return appVersionVisibleTeams, err
}

func (a AppVersionVisibleTeam) CreateAppVersionVisibleTeam(ct context.Context, appVersionVisibleTeam entity.AppVersionVisibleTeam) error {
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
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func (a AppVersionVisibleTeam) DeleteAppVersionVisibleTeam(ct context.Context, appID uint64, versionNumber int32, teamID uint64) error {
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
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: err})
	}

	return err
}

func NewAppVersionVisibleTeam(dataCollector telemetry.DataCollector, sqlDB *sql.DB) AppVersionVisibleTeam {
	return AppVersionVisibleTeam{dataCollector: dataCollector, db: sqlDB}
}
