package sqldb

import (
	"context"

	"github.com/teamyapp/cloud/libs/errs"
	"github.com/teamyapp/cloud/libs/transaction"
	"github.com/teamyapp/teamy-backend/core/dao"
	"github.com/teamyapp/teamy-backend/core/dao/entity"
)

const teamMemberGroupUserRelationDaoName = "TeamMemberGroupUserRelation"

type TeamMemberGroupUserRelation struct {
	metrics dao.Metrics
}

var _ dao.TeamMemberGroupUserRelation = TeamMemberGroupUserRelation{}

func (t TeamMemberGroupUserRelation) FindMemberGroupUserIDsByMemberGroupID(
	ct context.Context,
	tx *transaction.Transaction,
	memberGroupID uint64,
) ([]uint64, *errs.Error) {
	t.metrics.ReportDaoOperation(teamMemberGroupUserRelationDaoName, "FindMemberGroupUserIDsByMemberGroupID")
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
	t.metrics.ReportDaoOperation(teamMemberGroupUserRelationDaoName, "FindMemberGroupIDsByUserID")
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
	t.metrics.ReportDaoOperation(teamMemberGroupUserRelationDaoName, "CreateMemberGroupUserRelation")
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
	t.metrics.ReportDaoOperation(teamMemberGroupUserRelationDaoName, "DeleteMemberGroupUserRelation")
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

func (t TeamMemberGroupUserRelation) DeleteTeamMemberGroupUserRelationsByGroupID(
	ct context.Context,
	tx *transaction.Transaction,
	groupID uint64,
) *errs.Error {
	t.metrics.ReportDaoOperation(teamMemberGroupUserRelationDaoName, "DeleteTeamMemberGroupUserRelationsByGroupID")
	statement := `
		DELETE FROM team_member_group_user_relation
		WHERE group_id = $1;
	`
	_, err := tx.SQLTx().Exec(
		statement,
		groupID,
	)
	if err != nil {
		return errs.NewError(errs.Unknown, err.Error())
	}

	return nil
}

func NewTeamMemberGroupUserRelation(metrics dao.Metrics) TeamMemberGroupUserRelation {
	return TeamMemberGroupUserRelation{
		metrics: metrics,
	}
}
