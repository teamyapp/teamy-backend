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

type AppTeamInstallation struct {
	dataCollector telemetry.DataCollector
	db            *sql.DB
}

var _ dao.AppTeamInstallation = (*AppTeamInstallation)(nil)

func (a AppTeamInstallation) FindAppTeamInstallationsByAppID(ct context.Context, appID uint64) ([]entity.AppTeamInstallation, *errs.Error) {
	rows, err := a.db.Query(`
	SELECT
		app_id,
		installed_team_id,
		installed_by_user_id,
		enabled_version_number,
		installed_at
	FROM app_team_installation
	WHERE app_id = $1;
`,
		appID)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	appTeamInstallations := make([]entity.AppTeamInstallation, 0)
	for rows.Next() {
		appTeamInstallation := entity.AppTeamInstallation{}
		err = rows.Scan(
			&appTeamInstallation.AppID,
			&appTeamInstallation.InstalledTeamID,
			&appTeamInstallation.InstalledByUserID,
			&appTeamInstallation.EnabledVersionNumber,
			&appTeamInstallation.InstalledAt,
		)
		if err != nil {
			internalErr = &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{
				telemetry.CauseProp: internalErr,
			})
			continue
		}

		appTeamInstallations = append(appTeamInstallations, appTeamInstallation)
	}

	if internalErr != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	return appTeamInstallations, nil
}

func (a AppTeamInstallation) FindAppTeamInstallationsByTeamID(ct context.Context, teamID uint64) ([]entity.AppTeamInstallation, *errs.Error) {
	rows, err := a.db.Query(`
	SELECT
		app_id,
		installed_team_id,
		installed_by_user_id,
		enabled_version_number,
		installed_at
	FROM app_team_installation
	WHERE installed_team_id = $1;
`,
		teamID)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	defer rows.Close()

	var internalErr *errs.Error
	appTeamInstallations := make([]entity.AppTeamInstallation, 0)
	for rows.Next() {
		appTeamInstallation := entity.AppTeamInstallation{}
		err = rows.Scan(
			&appTeamInstallation.AppID,
			&appTeamInstallation.InstalledTeamID,
			&appTeamInstallation.InstalledByUserID,
			&appTeamInstallation.EnabledVersionNumber,
			&appTeamInstallation.InstalledAt,
		)
		if err != nil {
			internalErr = &errs.Error{
				Code:     errs.Unknown,
				EmbedErr: err,
			}
			a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{
				telemetry.CauseProp: internalErr,
			})
			continue
		}

		appTeamInstallations = append(appTeamInstallations, appTeamInstallation)
	}

	if internalErr != nil {
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return nil, internalErr
	}

	return appTeamInstallations, nil
}

func (a AppTeamInstallation) FindAppTeamInstallationByAppIDAndTeamID(ct context.Context, appID uint64, teamID uint64) (entity.AppTeamInstallation, *errs.Error) {
	appTeamInstallation := entity.AppTeamInstallation{}
	err := a.db.QueryRow(`
	SELECT
	    app_id,
	    installed_team_id,
	    installed_by_user_id,
	    enabled_version_number,
	    installed_at
	FROM app_team_installation
	WHERE app_id = $1 AND installed_team_id = $2;
`,
		appID,
		teamID).
		Scan(
			&appTeamInstallation.AppID,
			&appTeamInstallation.InstalledTeamID,
			&appTeamInstallation.InstalledByUserID,
			&appTeamInstallation.EnabledVersionNumber,
			&appTeamInstallation.InstalledAt,
		)

	if errors.Is(err, sql.ErrNoRows) {
		internalErr := &errs.Error{
			Code:    errs.NotFound,
			Message: fmt.Sprintf("app installation not found: appID=%v, teamID=%v", appID, teamID),
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.AppTeamInstallation{}, internalErr
	}

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return entity.AppTeamInstallation{}, internalErr
	}

	return appTeamInstallation, nil
}

func (a AppTeamInstallation) CreateAppTeamInstallation(ct context.Context, appTeamInstallation entity.AppTeamInstallation) *errs.Error {
	_, err := a.db.Exec(`
	INSERT INTO app_team_installation
	(
	 	app_id,
		installed_team_id,
	 	installed_by_user_id,
	 	enabled_version_number,
	 	installed_at
	)
	VALUES ($1, $2, $3, $4, $5);
`,
		appTeamInstallation.AppID,
		appTeamInstallation.InstalledTeamID,
		appTeamInstallation.InstalledByUserID,
		appTeamInstallation.EnabledVersionNumber,
		appTeamInstallation.InstalledAt,
	)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func (a AppTeamInstallation) UpdateAppTeamInstallation(ct context.Context, appTeamInstallation entity.AppTeamInstallation) *errs.Error {
	_, err := a.db.Exec(`
		UPDATE app_team_installation
		SET
			enabled_version_number = $1
		WHERE app_id = $2 AND installed_team_id = $3;`,
		appTeamInstallation.EnabledVersionNumber,
		appTeamInstallation.AppID,
		appTeamInstallation.InstalledTeamID,
	)

	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func (a AppTeamInstallation) DeleteAppTeamInstallation(ct context.Context, appID uint64, teamID uint64) *errs.Error {
	_, err := a.db.Exec(`
		DELETE FROM app_team_installation
		WHERE app_id = $1
		AND installed_team_id = $2;
		`,
		appID,
		teamID)
	if err != nil {
		internalErr := &errs.Error{
			Code:     errs.Unknown,
			EmbedErr: err,
		}
		a.dataCollector.Logger.LogWithContext(ct, telemetry.Error, telemetry.Props{telemetry.CauseProp: internalErr})
		return internalErr
	}

	return nil
}

func NewAppTeamInstallation(dataCollector telemetry.DataCollector, sqlDB *sql.DB) AppTeamInstallation {
	return AppTeamInstallation{dataCollector: dataCollector, db: sqlDB}
}
