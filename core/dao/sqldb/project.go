package sqldb

import (
	"context"
	"fmt"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/entity"
)

const projectDaoName = "Project"

type Project struct {
	metrics            dao.Metrics
	transactionFactory transaction.Factory
}

var _ dao.Project = (*Project)(nil)

func (p *Project) FindProjectsWithTx(ct context.Context, tx *transaction.Transaction) ([]entity.Project, *errs.Error) {
	p.metrics.ReportDaoOperation(projectDaoName, "FindProjectsWithTx")
	rows, err := tx.SQLTx().QueryContext(ct, `
		SELECT
			id,
			name,
			expected_start_at,
			actual_start_at,
			expected_end_at,
			actual_end_at,
			creator_id,
			created_at,
			updated_at,
			icon_url,
			color,
			team_id
		FROM project;
	`)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	var projects []entity.Project
	for rows.Next() {
		project := entity.Project{}
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.ExpectedStartAt,
			&project.ActualStartAt,
			&project.ExpectedEndAt,
			&project.ActualEndAt,
			&project.CreatorID,
			&project.CreatedAt,
			&project.UpdatedAt,
			&project.IconURL,
			&project.Color,
			&project.TeamID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		projects = append(projects, project)
	}

	return projects, nil
}

func (p *Project) FindProjectsByTeamIDWithTx(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.Project, *errs.Error) {
	p.metrics.ReportDaoOperation(projectDaoName, "FindProjectsByTeamIDWithTx")
	rows, err := tx.SQLTx().QueryContext(ct, `
		SELECT
			id,
			name,
			expected_start_at,
			actual_start_at,
			expected_end_at,
			actual_end_at,
			creator_id,
			created_at,
			updated_at,
			icon_url,
			color,
			team_id
			FROM project
		WHERE team_id = $1;
	`, teamID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	var projects []entity.Project
	for rows.Next() {
		project := entity.Project{}
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.ExpectedStartAt,
			&project.ActualStartAt,
			&project.ExpectedEndAt,
			&project.ActualEndAt,
			&project.CreatorID,
			&project.CreatedAt,
			&project.UpdatedAt,
			&project.IconURL,
			&project.Color,
			&project.TeamID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		projects = append(projects, project)
	}

	return projects, nil
}

func (p *Project) FindProjectsByIDsWithTx(ct context.Context, tx *transaction.Transaction, projectIDs []uint64) ([]entity.Project, *errs.Error) {
	p.metrics.ReportDaoOperation(projectDaoName, "FindProjectsByIDsWithTx")
	if len(projectIDs) == 0 {
		return []entity.Project{}, nil
	}

	projects := []entity.Project{}
	idsStr := toIDsString(projectIDs)
	query := fmt.Sprintf(`
		SELECT
			id,
			name,
			expected_start_at,
			actual_start_at,
			expected_end_at,
			actual_end_at,
			creator_id,
			created_at,
			updated_at,
			icon_url,
			color,
			team_id
		FROM project
		WHERE id IN (%s)
	`, idsStr)
	rows, err := tx.SQLTx().QueryContext(ct, query)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()
	for rows.Next() {
		project := entity.Project{}
		err := rows.Scan(
			&project.ID,
			&project.Name,
			&project.ExpectedStartAt,
			&project.ActualStartAt,
			&project.ExpectedEndAt,
			&project.ActualEndAt,
			&project.CreatorID,
			&project.CreatedAt,
			&project.UpdatedAt,
			&project.IconURL,
			&project.Color,
			&project.TeamID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		projects = append(projects, project)
	}

	return projects, nil
}

func (p *Project) FindProjectByIDWithTx(ct context.Context, tx *transaction.Transaction, projectID uint64) (entity.Project, *errs.Error) {
	p.metrics.ReportDaoOperation(projectDaoName, "FindProjectByIDWithTx")
	project := entity.Project{}
	err := tx.SQLTx().QueryRowContext(ct, `
		SELECT
			id,
			name,
			expected_start_at,
			actual_start_at,
			expected_end_at,
			actual_end_at,
			creator_id,
			created_at,
			updated_at,
			icon_url,
			color,
			team_id
		FROM project
		WHERE id = $1
	`, projectID).Scan(
		&project.ID,
		&project.Name,
		&project.ExpectedStartAt,
		&project.ActualStartAt,
		&project.ExpectedEndAt,
		&project.ActualEndAt,
		&project.CreatorID,
		&project.CreatedAt,
		&project.UpdatedAt,
		&project.IconURL,
		&project.Color,
		&project.TeamID,
	)

	if err != nil {
		return entity.Project{}, errs.NewError(errs.Unknown, err.Error())
	}

	return project, nil

}

func (p *Project) CreateProject(ct context.Context, tx *transaction.Transaction, project entity.Project) *errs.Error {
	p.metrics.ReportDaoOperation(projectDaoName, "CreateProject")
	_, err := tx.SQLTx().ExecContext(ct, `
		INSERT INTO project (
			id,
			name,
			expected_start_at,
			actual_start_at,
			expected_end_at,
			actual_end_at,
			creator_id,
			created_at,
			updated_at,
			icon_url,
			color,
			team_id
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);
	`,
		project.ID,
		project.Name,
		project.ExpectedStartAt,
		project.ActualStartAt,
		project.ExpectedEndAt,
		project.ActualEndAt,
		project.CreatorID,
		project.CreatedAt,
		project.UpdatedAt,
		project.IconURL,
		project.Color,
		project.TeamID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (p *Project) UpdateProject(ct context.Context, tx *transaction.Transaction, project entity.Project) *errs.Error {
	p.metrics.ReportDaoOperation(projectDaoName, "UpdateProject")
	_, err := tx.SQLTx().ExecContext(ct, `
		UPDATE project
		SET
			name = $1,
			expected_start_at = $2,
			actual_start_at = $3,
			expected_end_at = $4,
			actual_end_at = $5,
			creator_id = $6,
			created_at = $7,
			updated_at = $8,
			icon_url = $9,
			color = $10,
			team_id = $11
		WHERE id = $12;
	`,
		project.Name,
		project.ExpectedStartAt,
		project.ActualStartAt,
		project.ExpectedEndAt,
		project.ActualEndAt,
		project.CreatorID,
		project.CreatedAt,
		project.UpdatedAt,
		project.IconURL,
		project.Color,
		project.TeamID,
		project.ID,
	)

	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (p *Project) DeleteProject(ct context.Context, tx *transaction.Transaction, projectID uint64) *errs.Error {
	p.metrics.ReportDaoOperation(projectDaoName, "DeleteProject")
	_, err := tx.SQLTx().ExecContext(ct, `
		DELETE FROM project
		WHERE id = $1;
	`, projectID)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewProject(
	metrics dao.Metrics,
	transactionFactory transaction.Factory,
) *Project {
	return &Project{
		metrics:            metrics,
		transactionFactory: transactionFactory,
	}
}
