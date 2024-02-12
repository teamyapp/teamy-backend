package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/entity"
)

type TeamMemberGroupUserRelation struct {
}

var _ dao.TeamMemberGroupUserRelation = TeamMemberGroupUserRelation{}

func (t TeamMemberGroupUserRelation) FindMemberGroupUserIDsByMemberGroupID(
	ct context.Context,
	tx *transaction.Transaction,
	memberGroupID uint64,
) ([]uint64, *errs.Error) {
	query := `
		SELECT
			member_user_id
		FROM team_member_group_user_relation
		WHERE group_id = $1;
	`
	rows, err := tx.SQLTx().Query(query, memberGroupID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	var memberUserIDs []uint64
	for rows.Next() {
		var memberUserID uint64
		err = rows.Scan(
			&memberUserID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		memberUserIDs = append(memberUserIDs, memberUserID)
	}

	return memberUserIDs, nil
}

func (t TeamMemberGroupUserRelation) FindMemberGroupIDsByUserID(ct context.Context, tx *transaction.Transaction, userID uint64) ([]uint64, *errs.Error) {
	query := `
		SELECT
			group_id
		FROM team_member_group_user_relation
		WHERE member_user_id = $1;
	`
	rows, err := tx.SQLTx().Query(query, userID)
	if err != nil {
		return nil, errs.NewError(errs.Unknown, err.Error())
	}

	defer rows.Close()

	var memberGroupIDs []uint64
	for rows.Next() {
		var memberGroupID uint64
		err = rows.Scan(
			&memberGroupID,
		)
		if err != nil {
			return nil, errs.NewError(errs.Unknown, err.Error())
		}

		memberGroupIDs = append(memberGroupIDs, memberGroupID)
	}

	return memberGroupIDs, nil
}

func (t TeamMemberGroupUserRelation) CreateMemberGroupUserRelation(
	ct context.Context,
	tx *transaction.Transaction,
	relation entity.TeamMemberGroupUserRelation,
) *errs.Error {
	statement := `
		INSERT INTO team_member_group_user_relation
			(
			 group_id,
			 member_user_id,
			 created_at
			 )
		VALUES ($1, $2, $3);
	`
	_, err := tx.SQLTx().Exec(
		statement,
		relation.GroupID,
		relation.UserID,
		relation.CreatedAt,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func (t TeamMemberGroupUserRelation) DeleteMemberGroupUserRelation(
	ct context.Context,
	tx *transaction.Transaction,
	relation entity.TeamMemberGroupUserRelation,
) *errs.Error {
	statement := `
		DELETE FROM team_member_group_user_relation
		WHERE group_id = $1 AND member_user_id = $2;
	`
	_, err := tx.SQLTx().Exec(
		statement,
		relation.GroupID,
		relation.UserID,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTeamMemberGroupUserRelation() TeamMemberGroupUserRelation {
	return TeamMemberGroupUserRelation{}
}
