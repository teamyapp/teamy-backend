package sqldb

import (
	"context"
	"database/sql"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamAppInstallation struct {
	transactionFactory transaction.Factory
}

var _ dao.TeamAppInstallation = (*TeamAppInstallation)(nil)

func (*TeamAppInstallation) FindTeamAppInstallationByIDWithTx(ct context.Context, tx *transaction.Transaction, appInstallationID uint64) (entity.TeamAppInstallation, *errs.Error) {
	teamAppInstallation := entity.TeamAppInstallation{}
	err := tx.SQLTx().QueryRowContext(
		ct,
		`SELECT
			id,
			installed_team_id,
			app_id
			FROM team_app_installation
			WHERE id = $1`,
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

func (a *TeamAppInstallation) FindTeamAppInstallationsByAppIDWithTx(
	ct context.Context,
	tx *transaction.Transaction,
	appID uint64,
) ([]entity.TeamAppInstallation, *errs.Error) {
	teamAppInstallations := []entity.TeamAppInstallation{}
	rows, err := tx.SQLTx().QueryContext(
		ct,
		`SELECT
			id,
			installed_team_id,
			app_id
			FROM team_app_installation
			WHERE app_id = $1`,
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
func (*TeamAppInstallation) CreateTeamAppInstallation(
	ct context.Context,
	tx *transaction.Transaction,
	teamAppInstallation entity.TeamAppInstallation,
) *errs.Error {
	_, err := tx.SQLTx().ExecContext(
		ct,
		`INSERT INTO team_app_installation (
			id,
			installed_team_id,
			app_id
		) VALUES (
			$1,
			$2,
			$3
		)`,
		teamAppInstallation.ID,
		teamAppInstallation.InstalledTeamID,
		teamAppInstallation.AppID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (*TeamAppInstallation) DeleteTeamAppInstallationByID(ct context.Context, tx *transaction.Transaction, appInstallationID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(
		ct,
		`DELETE FROM team_app_installation
			WHERE id = $1`,
		appInstallationID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTeamAppInstallation(
	transactionFactory transaction.Factory,
) *TeamAppInstallation {
	return &TeamAppInstallation{
		transactionFactory: transactionFactory,
	}
}
