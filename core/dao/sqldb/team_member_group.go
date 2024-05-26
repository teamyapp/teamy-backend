package sqldb

import (
	"context"
	"database/sql"
	"errors"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/entity"
)

type TeamMemberGroup struct {
	metrics dao.Metrics
}

var _ dao.TeamMemberGroup = TeamMemberGroup{}

func (t TeamMemberGroup) FindMemberGroupByID(ct context.Context, tx *transaction.Transaction, id uint64) (entity.TeamMemberGroup, *errs.Error) {
	t.metrics.ReportDaoOperation("TeamMemberGroup", "FindMemberGroupByID")
	query := `
		SELECT
			id,
			team_id,
			name,
			order,
			authorization_user_group_id,
			created_at,
			updated_at
		FROM team_member_group
		WHERE id = $1`
	teamMemberGroup := entity.TeamMemberGroup{}
	err := tx.SQLTx().QueryRow(query, id).
		Scan(
			&teamMemberGroup.ID,
			&teamMemberGroup.TeamID,
			&teamMemberGroup.Name,
			&teamMemberGroup.Order,
			&teamMemberGroup.AuthorizationUserGroupID,
			&teamMemberGroup.CreatedAt,
			&teamMemberGroup.UpdatedAt,
		)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entity.TeamMemberGroup{}, errs.NewError(errs.NotFound, "team member group not found")
		}

		return entity.TeamMemberGroup{}, errs.NewError(errs.Unknown, err.Error())
	}

	return teamMemberGroup, nil
}

func (t TeamMemberGroup) FindMemberGroupsByTeamID(ct context.Context, tx *transaction.Transaction, teamID uint64) ([]entity.TeamMemberGroup, *errs.Error) {
	t.metrics.ReportDaoOperation("TeamMemberGroup", "FindMemberGroupsByTeamID")
	query := `
		SELECT
			id,
			team_id,
			name,
			order,
			authorization_user_group_id,
			created_at,
			updated_at
		FROM team_member_group
		WHERE team_id = $1`
	rows, err := tx.SQLTx().Query(query, teamID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	var memberGroups []entity.TeamMemberGroup
	for rows.Next() {
		memberGroup := entity.TeamMemberGroup{}
		err := rows.Scan(
			&memberGroup.ID,
			&memberGroup.TeamID,
			&memberGroup.Name,
			&memberGroup.Order,
			&memberGroup.AuthorizationUserGroupID,
			&memberGroup.CreatedAt,
			&memberGroup.UpdatedAt,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		memberGroups = append(memberGroups, memberGroup)
	}

	return memberGroups, nil
}

func (t TeamMemberGroup) CreateMemberGroup(ct context.Context, tx *transaction.Transaction, memberGroup entity.TeamMemberGroup) *errs.Error {
	t.metrics.ReportDaoOperation("TeamMemberGroup", "CreateMemberGroup")
	statement := `
		INSERT INTO team_member_group
			(
			 id,
			 team_id,
			 name,
			 order,
			 authorization_user_group_id,
			 created_at,
			 updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7)`
	_, err := tx.SQLTx().Exec(
		statement,
		memberGroup.ID,
		memberGroup.TeamID,
		memberGroup.Name,
		memberGroup.Order,
		memberGroup.AuthorizationUserGroupID,
		memberGroup.CreatedAt,
		memberGroup.UpdatedAt,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t TeamMemberGroup) UpdateMemberGroup(ct context.Context, tx *transaction.Transaction, memberGroup entity.TeamMemberGroup) *errs.Error {
	t.metrics.ReportDaoOperation("TeamMemberGroup", "UpdateMemberGroup")
	statement := `
		UPDATE team_member_group
		SET
			name = $1,
			team_id = $2,
			order = $3,
			authorization_user_group_id = $4,
			created_at = $5,
			updated_at = $6
		WHERE id = $7`
	_, err := tx.SQLTx().Exec(
		statement,
		memberGroup.Name,
		memberGroup.TeamID,
		memberGroup.Order,
		memberGroup.AuthorizationUserGroupID,
		memberGroup.CreatedAt,
		memberGroup.UpdatedAt,
		memberGroup.ID,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t TeamMemberGroup) DeleteMemberGroup(ct context.Context, tx *transaction.Transaction, id uint64) *errs.Error {
	t.metrics.ReportDaoOperation("TeamMemberGroup", "DeleteMemberGroup")
	statement := `
		DELETE FROM team_member_group
		WHERE id = $1`
	_, err := tx.SQLTx().Exec(statement, id)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTeamMemberGroup(metrics dao.Metrics) TeamMemberGroup {
	return TeamMemberGroup{
		metrics: metrics,
	}
}
