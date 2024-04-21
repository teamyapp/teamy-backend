package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const projectPhaseRelationDaoName = "ProjectPhaseRelation"

type ProjectPhaseRelation struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.ProjectPhaseRelation = (*ProjectPhaseRelation)(nil)

func (p *ProjectPhaseRelation) FindPhaseIDsByProjectIDWithTx(ct context.Context, tx *transaction.Transaction, projectID uint64) ([]uint64, *errs.Error) {
	p.metrics.ReportDaoOperation(projectPhaseRelationDaoName, "FindPhaseIDsByProjectIDWithTx")
	var phaseIDs []uint64
	rows, err := tx.SQLTx().QueryContext(ct, `
		SELECT
			phase_id
		FROM project_phase_relation
		WHERE project_id = $1
	`, projectID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		var phaseID uint64
		err := rows.Scan(&phaseID)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		phaseIDs = append(phaseIDs, phaseID)
	}

	return phaseIDs, nil
}

func (p *ProjectPhaseRelation) CreateProjectPhaseRelation(ct context.Context, tx *transaction.Transaction, projectPhaseRelation entity.ProjectPhaseRelation) *errs.Error {
	p.metrics.ReportDaoOperation(projectPhaseRelationDaoName, "CreateProjectPhaseRelation")
	_, err := tx.SQLTx().ExecContext(ct, `
		INSERT INTO project_phase_relation (
			project_id,
			phase_id
		) VALUES (
			$1,
			$2
		)
	`,
		projectPhaseRelation.ProjectID,
		projectPhaseRelation.PhaseID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (p *ProjectPhaseRelation) DeleteProjectPhaseRelation(ct context.Context, tx *transaction.Transaction, projectID uint64, phaseID uint64) *errs.Error {
	p.metrics.ReportDaoOperation(projectPhaseRelationDaoName, "DeleteProjectPhaseRelation")
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM project_phase_relation
		WHERE project_id = $1 AND phase_id = $2
	`, projectID, phaseID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (p *ProjectPhaseRelation) DeleteProjectPhaseRelationsByProjectID(ct context.Context, tx *transaction.Transaction, projectID uint64) *errs.Error {
	p.metrics.ReportDaoOperation(projectPhaseRelationDaoName, "DeleteProjectPhaseRelationsByProjectID")
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM project_phase_relation
		WHERE project_id = $1
	`, projectID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (p *ProjectPhaseRelation) DeleteProjectPhaseRelationsByPhaseID(ct context.Context, tx *transaction.Transaction, phaseID uint64) *errs.Error {
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM project_phase_relation
		WHERE phase_id = $1
	`, phaseID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewProjectPhaseRelation(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *ProjectPhaseRelation {
	return &ProjectPhaseRelation{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
