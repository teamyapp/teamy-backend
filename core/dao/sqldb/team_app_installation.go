package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const teamAppInstallationDaoName = "TeamAppInstallation"

type TeamAppInstallation struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.TeamAppInstallation = (*TeamAppInstallation)(nil)

func (t *TeamAppInstallation) FindTeamAppInstallationByIDWithTx(ct context.Context, tx *transaction.Transaction, appInstallationID uint64) (entity.TeamAppInstallation, *errs.Error) {
	t.metrics.ReportDaoOperation(teamAppInstallationDaoName, "FindTeamAppInstallationByIDWithTx")
	teamAppInstallation := entity.TeamAppInstallation{}
	err := tx.SQLTx().QueryRowContext(
		ct,
		`
		SELECT
			id,
			installed_team_id,
			app_id
		FROM team_app_installation
		WHERE id = $1;`,
		appInstallationID,
	).Scan(
		&teamAppInstallation.ID,
		&teamAppInstallation.InstalledTeamID,
		&teamAppInstallation.AppID,
	)

	if err != nil {
		return entity.TeamAppInstallation{}, errs.NewError(errs.Unknown, err.Error())
	}

	return teamAppInstallation, nil
}

func (t *TeamAppInstallation) FindTeamAppInstallationsByAppIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	appID uint64,
) ([]entity.TeamAppInstallation, *errs.Error) {
	t.metrics.ReportDaoOperation(teamAppInstallationDaoName, "FindTeamAppInstallationsByAppIDWithTx")
	var teamAppInstallations []entity.TeamAppInstallation
	rows, err := tx.SQLTx().QueryContext(
		ct,
		`
		SELECT
			id,
			installed_team_id,
			app_id
		FROM team_app_installation
		WHERE app_id = $1;`,
		appID,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		var teamAppInstallation entity.TeamAppInstallation
		err := rows.Scan(
			&teamAppInstallation.ID,
			&teamAppInstallation.InstalledTeamID,
			&teamAppInstallation.AppID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		teamAppInstallations = append(teamAppInstallations, teamAppInstallation)
	}

	return teamAppInstallations, nil
}

func (t *TeamAppInstallation) FindTeamAppInstallationsByAppID(ct context.Context, appID uint64) ([]entity.TeamAppInstallation, *errs.Error) {
	t.metrics.ReportDaoOperation(teamAppInstallationDaoName, "FindTeamAppInstallationsByAppID")
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := t.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return t.FindTeamAppInstallationsByAppIDWithTx(ct, tx, appID)
}
func (t *TeamAppInstallation) CreateTeamAppInstallation(
	ct context.Context,
	tx *transaction.Transaction,
	teamAppInstallation entity.TeamAppInstallation,
) *errs.Error {
	t.metrics.ReportDaoOperation(teamAppInstallationDaoName, "CreateTeamAppInstallation")
	_, err := tx.SQLTx().ExecContext(
		ct,
		`INSERT INTO team_app_installation (
			id,
			installed_team_id,
			app_id
		)
		VALUES (
			$1,
			$2,
			$3
		);`,
		teamAppInstallation.ID,
		teamAppInstallation.InstalledTeamID,
		teamAppInstallation.AppID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t *TeamAppInstallation) FindTeamAppInstallationsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.TeamAppInstallation, *errs.Error) {
	t.metrics.ReportDaoOperation(teamAppInstallationDaoName, "FindTeamAppInstallationsByTeamIDWithTx")
	var teamAppInstallations []entity.TeamAppInstallation
	rows, err := tx.SQLTx().QueryContext(
		ct,
		`
		SELECT
			id,
			installed_team_id,
			app_id
		FROM team_app_installation
		WHERE installed_team_id = $1;`,
		teamID,
	)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	for rows.Next() {
		var teamAppInstallation entity.TeamAppInstallation
		err := rows.Scan(
			&teamAppInstallation.ID,
			&teamAppInstallation.InstalledTeamID,
			&teamAppInstallation.AppID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		teamAppInstallations = append(teamAppInstallations, teamAppInstallation)
	}

	return teamAppInstallations, nil
}

func (t *TeamAppInstallation) FindTeamAppInstallationsByTeamID(ct context.Context, teamID uint64) ([]entity.TeamAppInstallation, *errs.Error) {
	t.metrics.ReportDaoOperation(teamAppInstallationDaoName, "FindTeamAppInstallationsByTeamID")
	opt := sql.TxOptions{
		ReadOnly: true,
	}
	tx, err := t.transactionFactory.BeginTx(ct, &opt)
	if err != nil {
		return nil, err
	}

	defer tx.Rollback()
	return t.FindTeamAppInstallationsByTeamIDWithTx(ct, tx, teamID)
}

func (t *TeamAppInstallation) DeleteTeamAppInstallationByID(ct context.Context, tx *transaction.Transaction, appInstallationID uint64) *errs.Error {
	t.metrics.ReportDaoOperation(teamAppInstallationDaoName, "DeleteTeamAppInstallationByID")
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
		DELETE FROM team_app_installation
		WHERE id = $1;`,
		appInstallationID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t *TeamAppInstallation) DeleteTeamAppInstallationsByAppID(ct context.Context, tx *transaction.Transaction, appID uint64) *errs.Error {
	t.metrics.ReportDaoOperation(teamAppInstallationDaoName, "DeleteTeamAppInstallationsByAppID")
	_, err := tx.SQLTx().ExecContext(
		ct,
		`
		DELETE FROM team_app_installation
		WHERE app_id = $1;`,
		appID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTeamAppInstallation(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *TeamAppInstallation {
	return &TeamAppInstallation{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
