package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

type TeamProjectRelation struct {
	transactionFactory transaction.Factory
}

var _ dao.TeamProjectRelation = (*TeamProjectRelation)(nil)

func (t *TeamProjectRelation) FindTeamIDsByProjectIDWithTx(ct context.Context, tx *transaction.Transaction, projectID uint64) ([]uint64, *errs.Error) {
	teamIDs := []uint64{}

	rows, err := tx.SQLTx().QueryContext(ct, `
		SELECT
			team_id
		FROM team_project_relation
		WHERE project_id = $1
	`, projectID)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		var teamID uint64
		err := rows.Scan(&teamID)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		teamIDs = append(teamIDs, teamID)
	}

	return teamIDs, nil
}

func (t *TeamProjectRelation) FindProjectIDsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]uint64, *errs.Error) {
	projectIDs := []uint64{}

	rows, err := tx.SQLTx().QueryContext(ct, `
		SELECT
			project_id
		FROM team_project_relation
		WHERE team_id = $1
	`, teamID)

	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		var projectID uint64
		err := rows.Scan(&projectID)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		projectIDs = append(projectIDs, projectID)
	}

	return projectIDs, nil
}

func (t *TeamProjectRelation) CreateTeamProjectRelation(ct context.Context, tx *transaction.Transaction, teamProjectRelation entity.TeamProjectRelation) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		INSERT INTO team_project_relation (
			team_id,
			project_id
		) VALUES (
			$1,
			$2
		)
	`,
		teamProjectRelation.TeamID,
		teamProjectRelation.ProjectID)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t *TeamProjectRelation) DeleteTeamProjectRelation(ct context.Context, tx *transaction.Transaction, teamID uint64, projectID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM team_project_relation
		WHERE team_id = $1 AND project_id = $2
	`, teamID, projectID)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTeamProjectRelation(transactionFactory transaction.Factory) *TeamProjectRelation {
	return &TeamProjectRelation{
		transactionFactory: transactionFactory,
	}
}
