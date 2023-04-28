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

type AppTeamInstallation struct {
	logger             telemetry.Logger
	transactionFactory transaction.Factory
}

var _ daov2.AppTeamInstallation = (*AppTeamInstallation)(nil)

func (a AppTeamInstallation) FindAppTeamInstallationsByAppID(ct context.Context,
	appID uint64) ([]entity.AppTeamInstallation, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return a.FindAppTeamInstallationsByAppIDWithTx(ct, tx, appID)
}

func (a AppTeamInstallation) FindAppTeamInstallationsByTeamID(ct context.Context,
	teamID uint64) ([]entity.AppTeamInstallation, *errs.Error) {
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := a.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return a.FindAppTeamInstallationsByTeamIDWithTx(ct, tx, teamID)
}

func (a AppTeamInstallation) FindAppTeamInstallationsByAppIDWithTx(ct context.Context, tx *transaction.Transaction,
	appID uint64) ([]entity.AppTeamInstallation, *errs.Error) {
	rows, err := tx.SQLTx().Query(`
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
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

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
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		appTeamInstallations = append(appTeamInstallations, appTeamInstallation)
	}

	return appTeamInstallations, nil
}

func (a AppTeamInstallation) FindAppTeamInstallationsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction,
	teamID uint64) ([]entity.AppTeamInstallation, *errs.Error) {
	rows, err := tx.SQLTx().Query(`
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
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

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
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		appTeamInstallations = append(appTeamInstallations, appTeamInstallation)
	}

	return appTeamInstallations, nil
}

func (a AppTeamInstallation) FindAppTeamInstallationByAppIDAndTeamIDWithTx(ct context.Context,
	tx *transaction.Transaction, appID uint64, teamID uint64) (entity.AppTeamInstallation, *errs.Error) {
	appTeamInstallation := entity.AppTeamInstallation{}
	err := tx.SQLTx().QueryRow(`
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

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.AppTeamInstallation{}, errs.NewError(errs.NotFound,
				fmt.Sprintf("app installation not found: appID=%v, teamID=%v", appID, teamID))
		}

		return entity.AppTeamInstallation{}, errs.NewError(errs.NotFound, err.Error())
	}

	return appTeamInstallation, nil
}

func (a AppTeamInstallation) CreateAppTeamInstallation(ct context.Context, tx *transaction.Transaction,
	appTeamInstallation entity.AppTeamInstallation) *errs.Error {
	_, err := tx.SQLTx().Exec(`
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
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a AppTeamInstallation) UpdateAppTeamInstallation(ct context.Context, tx *transaction.Transaction,
	appTeamInstallation entity.AppTeamInstallation) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		UPDATE app_team_installation
		SET
			enabled_version_number = $1
		WHERE app_id = $2 AND installed_team_id = $3;`,
		appTeamInstallation.EnabledVersionNumber,
		appTeamInstallation.AppID,
		appTeamInstallation.InstalledTeamID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a AppTeamInstallation) DeleteAppTeamInstallation(ct context.Context, tx *transaction.Transaction, appID uint64,
	teamID uint64) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM app_team_installation
		WHERE app_id = $1
		AND installed_team_id = $2;
		`,
		appID,
		teamID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (a AppTeamInstallation) DeleteAppTeamInstallationsByAppIDAndVersionNumber(ct context.Context,
	tx *transaction.Transaction, appID uint64, versionNumber int32) *errs.Error {
	_, err := tx.SQLTx().Exec(`
		DELETE FROM app_team_installation
		WHERE app_id = $1 AND enabled_version_number = $2;
		`,
		appID,
		versionNumber)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewAppTeamInstallation(logger telemetry.Logger, sqlDB *sql.DB) AppTeamInstallation {
	return AppTeamInstallation{logger: logger}
}
