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
}

var _ dao.TeamMemberGroup = TeamMemberGroup{}

func (t TeamMemberGroup) FindMemberGroupByID(ct context.Context, tx *transaction.Transaction, id uint64) (entity.TeamMemberGroup, *errs.Error) {
	query := `
		SELECT
			id,
			team_id,
			name,
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
			&teamMemberGroup.AuthorizationUserGroupID,
			&teamMemberGroup.CreatedAt,
			&teamMemberGroup.UpdatedAt,
			teamMemberGroup.ID,
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
	query := `
		SELECT
			id,
			team_id,
			name,
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
	statement := `
		INSERT INTO team_member_group
			(
			 id,
			 team_id,
			 name,
			 authorization_user_group_id,
			 created_at,
			 updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6)`
	_, err := tx.SQLTx().Exec(
		statement,
		memberGroup.ID,
		memberGroup.TeamID,
		memberGroup.Name,
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
	statement := `
		UPDATE team_member_group
		SET
			name = $1,
			team_id = $2,
			authorization_user_group_id = $3,
			created_at = $4,
			updated_at = $5
		WHERE id = $6`
	_, err := tx.SQLTx().Exec(
		statement,
		memberGroup.Name,
		memberGroup.TeamID,
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
	statement := `
		DELETE FROM team_member_group
		WHERE id = $1`
	_, err := tx.SQLTx().Exec(statement, id)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTeamMemberGroup() TeamMemberGroup {
	return TeamMemberGroup{}
}
